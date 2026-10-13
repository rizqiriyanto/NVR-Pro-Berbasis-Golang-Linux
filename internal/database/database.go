package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Camera adalah cetak biru (struktur data) untuk menyimpan data kamera di database
type Camera struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"unique;not null"`
	RTSPURL   string `gorm:"not null"`
	Status    string `gorm:"default:'offline'"` // Akan berisi: online, offline, atau recording
}

// SystemSetting adalah tabel baru untuk menyimpan ketikan lokasi hardisk sesuka hati Anda
type SystemSetting struct {
	ID          uint   `gorm:"primaryKey"`
	StoragePath string `gorm:"not null;default:'./records'"`
}

// DB adalah variabel global agar bisa dipanggil dari file manager atau main.go
var DB *gorm.DB

// InitDB bertugas membuat dan menghubungkan aplikasi ke file nvr.db
func InitDB() {
	var err error
	// Membuka koneksi ke file nvr.db (jika file belum ada, SQLite akan otomatis membuatnya)
	DB, err = gorm.Open(sqlite.Open("nvr.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal terhubung ke database SQLite:", err)
	}

	// AutoMigrate sekarang akan membuat 2 tabel: cameras dan system_settings
	err = DB.AutoMigrate(&Camera{}, &SystemSetting{})
	if err != nil {
		log.Fatal("Gagal membuat struktur tabel database:", err)
	}

	log.Println("✅ Database SQLite berhasil diinisialisasi (nvr.db)")
}
