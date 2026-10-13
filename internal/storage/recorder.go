package storage

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func StartRecording(camID uint, rtspURL string, baseStoragePath string) {
	camFolder := filepath.Join(baseStoragePath, fmt.Sprintf("cam_%d", camID))

	// Membuat direktori folder kamera jika belum ada
	if err := os.MkdirAll(camFolder, os.ModePerm); err != nil {
		log.Printf("Gagal membuat folder untuk Cam %d: %v\n", camID, err)
		return
	}

	log.Printf("🎥 [Cam %d] Memulai rekaman ke: %s\n", camID, camFolder)
	outputFormat := filepath.Join(camFolder, "vid-%Y%m%d-%H%M%S.mp4")

	for {
		// Menggunakan encoding ultrafast x264 untuk me-repair NAL unit corrupt/non-standard RTSP HP & CCTV
		cmd := exec.Command("ffmpeg",
			"-rtsp_transport", "tcp",
			"-analyzeduration", "10000000",
			"-probesize", "10000000",
			"-fflags", "+genpts+discardcorrupt",
			"-avoid_negative_ts", "make_zero",
			"-i", rtspURL,
			"-map", "0:v:0?",
			"-map", "0:a:0?",
			"-c:v", "libx264",
			"-preset", "ultrafast",
			"-tune", "zerolatency",
			"-c:a", "copy",
			"-f", "segment",
			"-segment_time", "60",
			"-reset_timestamps", "1",
			"-strftime", "1",
			"-movflags", "+faststart",
			outputFormat,
		)

		// Senyapkan log FFmpeg agar terminal tetap bersih dan hemat CPU
		cmd.Stderr = nil
		cmd.Stdout = nil

		err := cmd.Run()
		if err != nil {
			log.Printf("⚠️ [Cam %d] Perekaman terhenti. Mencoba konek ulang dalam 5 detik...\n", camID)
		}

		time.Sleep(5 * time.Second)
	}
}
