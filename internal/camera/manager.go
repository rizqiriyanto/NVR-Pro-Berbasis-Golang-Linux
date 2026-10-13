package camera

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"nvr-rumah/internal/config"
	"nvr-rumah/internal/database"
	"nvr-rumah/internal/storage"
)

func InitSystem(appConfig *config.AppConfig) string {
	var setting database.SystemSetting
	if err := database.DB.First(&setting).Error; err != nil {
		setting = database.SystemSetting{StoragePath: appConfig.StoragePath}
		database.DB.Create(&setting)
	}

	var count int64
	database.DB.Model(&database.Camera{}).Count(&count)
	if count == 0 {
		log.Println("Memasukkan kamera pertama ke database...")
		firstCam := database.Camera{
			Name:    appConfig.CameraName,
			RTSPURL: appConfig.CameraRTSPURL,
			Status:  "recording",
		}
		database.DB.Create(&firstCam)
	}

	return setting.StoragePath
}

func StartSemuaKamera(activeStoragePath string) {
	var cameras []database.Camera
	database.DB.Find(&cameras)

	log.Printf("Mendeteksi %d kamera di database.\n", len(cameras))

	for _, cam := range cameras {
		if cam.Status != "offline" {
			log.Printf("▶️ Menjalankan kamera: %s\n", cam.Name)
			
			// 1. Perekam ke Hardisk (Storage Package)
			go storage.StartRecording(cam.ID, cam.RTSPURL, activeStoragePath)

			// 2. Streamer ke RAM/Web Folder untuk Live View (Camera Package)
			go StartHLS(cam.ID, cam.RTSPURL)
		}
	}
}

// GetRecordsByFolder bertugas membaca daftar file .mp4 dari sub-folder kamera
func GetRecordsByFolder(camFolder string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(camFolder)
	if err != nil {
		return nil, err
	}

	folderName := filepath.Base(camFolder)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".mp4" {
			// Format yang dikembalikan: cam_1/vid-20261012-142821.mp4
			files = append(files, fmt.Sprintf("%s/%s", folderName, entry.Name()))
		}
	}

	return files, nil
}
