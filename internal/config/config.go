package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// AppConfig adalah cetakan/struktur yang akan menampung semua pengaturan kita
type AppConfig struct {
	ServerPort    string
	CameraName    string
	CameraRTSPURL string
	StoragePath   string // <-- INI BARIS YANG MEMBUAT ERROR SEBELUMNYA
}

// LoadConfig bertugas membaca file .env dan memindahkannya ke dalam AppConfig
func LoadConfig() *AppConfig {
	err := godotenv.Load()
	if err != nil {
		log.Println("Info: File .env tidak ditemukan. Menggunakan nilai bawaan.")
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = ":8088"
	}

	storage := os.Getenv("STORAGE_PATH")
	if storage == "" {
		storage = "./records" // Jika tidak diisi, gunakan folder bawaan
	}

	return &AppConfig{
		ServerPort:    port,
		CameraName:    os.Getenv("CAMERA_NAME"),
		CameraRTSPURL: os.Getenv("CAMERA_RTSP_URL"),
		StoragePath:   storage,
	}
}
