package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"nvr-rumah/internal/camera"
	"nvr-rumah/internal/config"
	"nvr-rumah/internal/database"
	"nvr-rumah/internal/storage"
)

func main() {
	appConfig := config.LoadConfig()

	// 1. Inisialisasi Database SQLite
	database.InitDB()

	// 2. Inisialisasi System Setting & Jalankan Semua Kamera
	activeStoragePath := camera.InitSystem(appConfig)
	go camera.StartSemuaKamera(activeStoragePath)

	// 3. Konfigurasi Web Server & API
	
	// API GET: Melihat daftar kamera
	http.HandleFunc("/api/cameras", func(w http.ResponseWriter, r *http.Request) {
		var cameras []database.Camera
		database.DB.Find(&cameras)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cameras)
	})

	// API POST: Menambah kamera baru
	http.HandleFunc("/api/cameras/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var newCam database.Camera
		if err := json.NewDecoder(r.Body).Decode(&newCam); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		newCam.Status = "recording"
		if err := database.DB.Create(&newCam).Error; err != nil {
			http.Error(w, "Gagal menyimpan kamera", http.StatusInternalServerError)
			return
		}

		// Baca storage path aktif saat ini dari database
		var setting database.SystemSetting
		database.DB.First(&setting)
		currentPath := setting.StoragePath
		if currentPath == "" {
			currentPath = "./records"
		}

		// Jalankan proses perekaman dan streamer live untuk kamera baru
		go storage.StartRecording(newCam.ID, newCam.RTSPURL, currentPath)
		go camera.StartHLS(newCam.ID, newCam.RTSPURL)
		
		log.Printf("➕ Kamera baru berhasil ditambahkan: %s (#%d)\n", newCam.Name, newCam.ID)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(newCam)
	})

	// API POST: Menghapus kamera dari database
	http.HandleFunc("/api/cameras/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ID uint `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if err := database.DB.Delete(&database.Camera{}, req.ID).Error; err != nil {
			http.Error(w, "Gagal menghapus kamera", http.StatusInternalServerError)
			return
		}

		log.Printf("🗑️ [System] Kamera #%d berhasil dihapus dari database\n", req.ID)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Kamera berhasil dihapus"})
	})

	// API GET: Membaca lokasi penyimpanan hardisk saat ini dari SQLite
	http.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) {
		var setting database.SystemSetting
		database.DB.First(&setting)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(setting)
	})

	// API POST: Memperbarui lokasi penyimpanan hardisk di database
	http.HandleFunc("/api/settings/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req database.SystemSetting
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.StoragePath == "" {
			http.Error(w, "Path penyimpanan tidak valid", http.StatusBadRequest)
			return
		}

		// Otomatis buat direktori baru di hardisk jika belum ada
		if err := os.MkdirAll(req.StoragePath, os.ModePerm); err != nil {
			log.Printf("⚠️ Gagal membuat direktori hardisk baru: %v\n", err)
		}

		var setting database.SystemSetting
		if err := database.DB.First(&setting).Error; err != nil {
			setting = database.SystemSetting{StoragePath: req.StoragePath}
			database.DB.Create(&setting)
		} else {
			setting.StoragePath = req.StoragePath
			database.DB.Save(&setting)
		}

		log.Printf("💾 [System] Path penyimpanan hardisk diperbarui ke: %s\n", req.StoragePath)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Path berhasil diperbarui"})
	})

	// API GET: Membaca daftar file rekaman berdasarkan kamera & lokasi storage dinamis
	http.HandleFunc("/api/records", func(w http.ResponseWriter, r *http.Request) {
		camIDStr := r.URL.Query().Get("cam_id")
		if camIDStr == "" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]string{})
			return
		}

		var setting database.SystemSetting
		database.DB.First(&setting)
		storagePath := setting.StoragePath
		if storagePath == "" {
			storagePath = appConfig.StoragePath
		}

		camFolder := filepath.Join(storagePath, fmt.Sprintf("cam_%s", camIDStr))
		files, err := camera.GetRecordsByFolder(camFolder)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]string{})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(files)
	})

	// Membuka akses static folder web (HTML, CSS, JS)
	http.Handle("/", http.FileServer(http.Dir("web")))

	// Membuka akses file rekaman arsip secara dinamis sesuai DB StoragePath saat ini
	http.HandleFunc("/records/", func(w http.ResponseWriter, r *http.Request) {
		var setting database.SystemSetting
		database.DB.First(&setting)
		storagePath := setting.StoragePath
		if storagePath == "" {
			storagePath = appConfig.StoragePath
		}
		http.StripPrefix("/records/", http.FileServer(http.Dir(storagePath))).ServeHTTP(w, r)
	})

	// Membuka akses folder siaran langsung (HLS Live Stream)
	http.Handle("/streams/", http.StripPrefix("/streams/", http.FileServer(http.Dir("web/streams"))))

	// Jalankan Web Server NVR
	fmt.Printf("\n🚀 Server NVR siap! Akses Web UI di port %s\n", appConfig.ServerPort)
	log.Fatal(http.ListenAndServe(appConfig.ServerPort, nil))
}
