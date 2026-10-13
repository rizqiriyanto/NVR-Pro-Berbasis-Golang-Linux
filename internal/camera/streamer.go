package camera

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// StartHLS bertugas menjalankan FFmpeg HLS dengan opsi probing & mapping yang stabil
func StartHLS(camID uint, rtspURL string) {
	streamFolder := filepath.Join(".", "web", "streams", fmt.Sprintf("cam_%d", camID))

	os.RemoveAll(streamFolder)
	if err := os.MkdirAll(streamFolder, os.ModePerm); err != nil {
		log.Printf("Gagal membuat folder stream untuk Cam %d: %v\n", camID, err)
		return
	}

	outputFile := filepath.Join(streamFolder, "index.m3u8")
	log.Printf("📡 [Cam %d] Memulai mesin Live View (HLS)...", camID)

	cmd := exec.Command("ffmpeg",
		"-rtsp_transport", "tcp",
		"-analyzeduration", "10000000",
		"-probesize", "10000000",
		"-fflags", "+genpts",
		"-i", rtspURL,
		"-map", "0:v:0?",
		"-c:v", "copy",
		"-an",
		"-f", "hls",
		"-hls_time", "2",
		"-hls_list_size", "3",
		"-hls_flags", "delete_segments",
		outputFile,
	)

	// Sembunyikan log mentah FFmpeg agar terminal bersih
	cmd.Stderr = nil
	cmd.Stdout = nil

	err := cmd.Run()
	if err != nil {
		log.Printf("❌ Mesin Live View terhenti untuk Cam %d\n", camID)
	}
}
