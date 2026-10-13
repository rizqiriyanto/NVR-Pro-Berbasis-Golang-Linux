package storage

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func StartAutoRotation(maxSizeMB int64, baseStoragePath string) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	maxSizeBytes := maxSizeMB * 1024 * 1024

	log.Printf("🧹 Rotasi Multi-Kamera Aktif. Memantau: %s (Batas: %d MB/kamera)\n", baseStoragePath, maxSizeMB)

	cleanUpAllCameras(baseStoragePath, maxSizeBytes)
	for range ticker.C {
		cleanUpAllCameras(baseStoragePath, maxSizeBytes)
	}
}

// PERUBAHAN STEP 3: Masuk ke dalam setiap sub-folder kamera
func cleanUpAllCameras(baseDir string, maxSize int64) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		// Jika ini adalah folder (misal cam_1)
		if entry.IsDir() {
			camPath := filepath.Join(baseDir, entry.Name())
			cleanUpDisk(camPath, maxSize)
		}
	}
}

func cleanUpDisk(dir string, maxSize int64) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	type fileData struct {
		path    string
		size    int64
		modTime time.Time
	}

	var files []fileData
	var totalSize int64

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".mp4" {
			continue
		}
		
		info, err := entry.Info()
		if err != nil {
			continue
		}

		size := info.Size()
		files = append(files, fileData{
			path:    filepath.Join(dir, entry.Name()),
			size:    size,
			modTime: info.ModTime(),
		})
		totalSize += size
	}

	if totalSize <= maxSize {
		return
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})

	for _, f := range files {
		if totalSize <= maxSize {
			break
		}
		
		err := os.Remove(f.path)
		if err == nil {
			log.Printf("[PEMBERSIHAN] Menghapus video lama %s\n", filepath.Base(f.path))
			totalSize -= f.size
		}
	}
}
