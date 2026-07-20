package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ballspins/environmental-data-logger/backend/benchmark/reporter"
)

func main() {
	log.Println("[Orchestrator] Memulai eksekusi seluruh skenario benchmark...")

	// 1. Buat direktori penyimpanan hasil bertimestamp (results/YYYY-MM-DD_HH-MM-SS/)
	currentTime := time.Now().Format("2006-01-02_15-04-05")
	outputDir := filepath.Join("results", currentTime)
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Fatalf("[Orchestrator] Gagal membuat direktori hasil: %v", err)
	}
	log.Printf("[Orchestrator] Hasil eksperimen akan disimpan di: %s\n", outputDir)

	// 2. Simpan metadata.json ke dalam direktori tersebut
	err = reporter.SaveMetadata(outputDir, "Redis 6.2+", "MySQL 8.0+")
	if err != nil {
		log.Printf("[Orchestrator] Gagal menyimpan file metadata: %v\n", err)
	}

	// 3. Daftarkan seluruh program benchmark yang akan dijalankan sekuensial
	benchmarks := []struct {
		name string
		args []string
	}{
		{
			name: "benchmark-json-binary",
			args: []string{"--count=3", "--duration=3s"},
		},
		{
			name: "benchmark-chunk",
			args: []string{"--count=3", "--duration=2s"},
		},
		{
			name: "benchmark-db",
			args: []string{"--count=3"},
		},
		{
			name: "benchmark-redis",
			args: []string{"--count=3", "--burst-rate=200"},
		},
		{
			name: "stress-shadow-device",
			args: []string{"--nodes=200", "--duration=10s", "--payload=binary", "--pattern=constant"},
		},
	}

	// 4. Eksekusi program sekuensial menggunakan go run
	for i, bench := range benchmarks {
		fmt.Printf("\n=======================================================\n")
		log.Printf("[%d/%d] MENJALANKAN: %s %v\n", i+1, len(benchmarks), bench.name, bench.args)
		fmt.Printf("=======================================================\n")

		// Gabungkan argumen default dengan target output folder bertimestamp
		args := append([]string{"run", fmt.Sprintf("./cmd/%s/main.go", bench.name)}, bench.args...)
		args = append(args, fmt.Sprintf("--output=%s", outputDir))

		cmd := exec.Command("go", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		start := time.Now()
		err := cmd.Run()
		if err != nil {
			log.Printf("[ERROR] Gagal mengeksekusi %s: %v\n", bench.name, err)
			continue
		}
		log.Printf("[SUCCESS] %s selesai dalam %v\n", bench.name, time.Since(start))
	}

	fmt.Printf("\n=======================================================\n")
	log.Println("[Orchestrator] Seluruh rangkaian pengujian benchmark SELESAI.")
	log.Printf("[Orchestrator] Laporan lengkap (CSV, JSON, MD) tersedia di: %s/\n", outputDir)
	fmt.Printf("=======================================================\n")
}
