# 📹 NVR Pro - Network Video Recorder Berbasis Golang & Linux

**NVR Pro** adalah aplikasi Network Video Recorder (NVR) CCTV rumahan berkinerja tinggi yang dibangun menggunakan **Golang**, **FFmpeg**, **SQLite**, dan **Tailwind CSS**. Aplikasi ini didesain agar sangat ringan di CPU, mendukung banyak kamera (*multi-camera*), serta mendukung penyimpanan kustom ke hardisk eksternal/NAS (CasaOS).

---

## 🌟 Fitur Utama

- **Live View Multi-Kamera (HLS Stream):** Pemantauan siaran langsung *real-time* dari berbagai IP Camera / RTSP HP dengan *delay* rendah (3-5 detik).
- **Perekaman 24/7 Kontinu:** Merekam dan memotong video otomatis per 60 detik dalam format `.mp4`.
- **Ultrafast Transcoding & Repair (Kamera HP & CCTV):** Mengatasi isu *Invalid NAL Unit Size / Frame Corrupt* dari aplikasi RTSP HP sehingga hasil rekaman di *Playback* dipastikan **100% muncul gambar & suara** tanpa layar hitam.
- **Custom Storage Path:** Bebas menentukan lokasi direktori penyimpanan rekaman (misal: `/mnt/hdd500g/records` atau lokasi hardisk CasaOS).
- **Database SQLite Mode WAL:** Mencegah terjadinya *locking/busy timeout* saat banyak kamera melakukan pembacaan dan penulisan data secara bersamaan.
- **Manajemen Kamera Dinamis:** Fitur Tambah, Edit Path, dan Hapus Kamera langsung melalui antarmuka Web UI tanpa restart manual.
- **Antarmuka Web Responsif (Mobile & Desktop):** Tampilan modern dengan navigasi *drawer* seluler dan *timeline player* rekaman arsip.
- **Systemd Background Service:** Berjalan 24 jam nonstop di belakang layar Linux dan otomatis menyala saat server reboot.

---

## 🛠️ Tech Stack & Arsitektur

- **Backend:** Go (Golang) 1.20+
- **Database:** SQLite3 (GORM Driver) dengan pragma `_journal_mode=WAL`
- **Video Engine:** FFmpeg (Segment Muxer, HLS Streamer, x264 Ultrafast Encoder)
- **Frontend:** HTML5, JavaScript Vanilla, Tailwind CSS (CDN), HLS.js
- **Sistem Operasi:** Linux / Ubuntu Server / CasaOS / Debian

---

## 📁 Struktur Proyek

```text
nvr-rumah/
├── cmd/
│   └── server/
│       └── main.go              # Server HTTP Utama & Endpoint API
├── internal/
│   ├── camera/
│   │   ├── manager.go           # Inisialisasi & Pengelola Daftar Kamera
│   │   └── streamer.go          # Mesin Live Streamer HLS (FFmpeg)
│   ├── config/
│   │   └── config.go            # Pembaca File Konfigurasi .env
│   ├── database/
│   │   └── database.go          # Handler Database SQLite WAL Mode
│   └── storage/
│       ├── recorder.go          # Mesin Perekam Video (FFmpeg Ultrafast Transcoder)
│       └── rotator.go           # Pembersih Otomatis File Rekaman Lama
├── web/
│   ├── index.html               # Web UI Dashboard Responsif
│   └── streams/                 # Folder Penyimpanan Sementara Live HLS (.m3u8 & .ts)
├── .env                         # Konfigurasi Awal Environment
├── go.mod                       # Modul Dependensi Golang
└── README.md                    # Dokumentasi Proyek
```

---

## 🚀 Panduan Instalasi & Pengoperasian (Step-by-Step)

### Step 1: Install Dependensi Sistem Linux
Pastikan **Golang** dan **FFmpeg** telah terpasang di server Linux Anda:

```bash
# Update repository
sudo apt update

# Install Golang dan FFmpeg
sudo apt install golang ffmpeg git -y

# Cek versi untuk memastikan instalasi berhasil
go version
ffmpeg -version
```

### Step 2: Clone & Inisialisasi Proyek
```bash
# Clone repository ini (atau buat folder jika belum di-push)
git clone https://github.com/rizqiriyanto/nvr-rumah.git
cd nvr-rumah

# Download dependensi Golang
go mod tidy
```

### Step 3: Uji Coba Mode Development
```bash
go run cmd/server/main.go
```
Buka browser Anda dan akses: `http://IP_SERVER_ANDA:8088` (Contoh: `http://192.168.2.5:8088`).

---

## ⚙️ Menjalankan di Belakang Layar (Systemd Background Service)

Agar NVR berjalan 24 jam nonstop tanpa perlu membuka terminal, ikuti langkah berikut:

### 1. Build Kode Golang Menjadi Binary Executable
```bash
cd /root/nvr-rumah
go build -o nvr-server cmd/server/main.go
```

### 2. Buat File Service Systemd
```bash
sudo nano /etc/systemd/system/nvr.service
```

Tempelkan konfigurasi berikut:
```ini
[Unit]
Description=NVR Pro CCTV Server
After=network.target network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=/root/nvr-rumah
ExecStart=/root/nvr-rumah/nvr-server
Restart=always
RestartSec=5
KillMode=process
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

### 3. Aktifkan & Jalankan Service
```bash
# Reload daemon systemd
sudo systemctl daemon-reload

# Aktifkan auto-start saat reboot
sudo systemctl enable nvr

# Jalankan service sekarang
sudo systemctl start nvr

# Cek status service
sudo systemctl status nvr
```

---

## 🛠️ Catatan Solusi Troubleshooting Teknis (Developer Log)

 Selama masa pengembangan, ditemukan beberapa kendala unik pada stream RTSP dan SQLite yang berhasil diselesaikan dengan metode berikut:

| Kendala Teknis | Penyebab Akar Masalah | Solusi Terapan |
| :--- | :--- | :--- |
| **Layar Hitam pada Playback HP** | Aplikasi RTSP HP mengirimkan NAL Unit corrupt (`Invalid NAL unit size`) serta tidak menyertakan SPS/PPS header di awal segmen MP4. | Menggunakan encoder `-c:v libx264 -preset ultrafast -tune zerolatency` pada `recorder.go` untuk menstandarkan struktur H.264 MP4. |
| **Database Locked / Busy Error** | SQLite bawaan terkunci saat proses *Insert* kamera dan penulisan status berjalan bersamaan dengan *Read* API. | Menambahkan Pragmas `_journal_mode=WAL` dan `_busy_timeout=5000` di `database.go`. |
| **Terminal Lag / CPU High** | FFmpeg mencetak jutaan baris log *warning* ke terminal Linux (`Stderr`). | Mengalihkan/menyelapkan `cmd.Stderr = nil` pada proses FFmpeg `recorder.go` dan `streamer.go`. |

---

## 📤 Panduan Upload Proyek ke GitHub

Ikuti langkah ini untuk mengunggah proyek ini ke akun GitHub Anda untuk pertama kali:

### 1. Buat File `.gitignore`
Buat file bernama `.gitignore` di folder utama `nvr-rumah` agar file rekaman video, database, dan file biner **tidak ikut terunggah** ke GitHub (menghemat kuota & menjaga privasi):

```bash
nano .gitignore
```

Isi dengan:
```text
# Binary file
nvr-server

# Database SQLite
*.db
*.db-journal
*.db-wal
*.db-shm

# Folder Rekaman & Stream
records/
web/streams/
*.mp4

# File Lingkungan (.env)
.env
```

### 2. Inisialisasi Git & Push ke GitHub
```bash
# Inisialisasi Git
git init

# Tambahkan semua file proyek
git add .

# Buat Commit pertama
git commit -m "Initial Commit: NVR Pro Golang Complete System"

# Hubungkan ke repository GitHub Anda (Ganti URL dengan repository GitHub Anda)
git branch -M main
git remote add origin https://github.com/USERNAME_ANDA/nvr-rumah.git

# Upload ke GitHub
git push -u origin main
```

---

## 📜 Lisensi
Dikembangkan untuk penggunaan pribadi dan edukasi open-source under the [MIT License](LICENSE).
