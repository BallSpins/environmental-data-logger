package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ballspins/environmental-data-logger/backend/benchmark/generator"
	"github.com/ballspins/environmental-data-logger/backend/benchmark/metrics"
	"github.com/ballspins/environmental-data-logger/backend/benchmark/profile"
	"github.com/ballspins/environmental-data-logger/backend/benchmark/reporter"
	"github.com/ballspins/environmental-data-logger/backend/internal/config"
	"github.com/ballspins/environmental-data-logger/backend/internal/model"
	"github.com/ballspins/environmental-data-logger/backend/internal/repository"
)

func main() {
	countFlag := flag.Int("count", 5, "Jumlah repetisi eksekusi benchmark")
	outputFlag := flag.String("output", "results", "Direktori penyimpanan output hasil")
	pprofFlag := flag.Bool("pprof", false, "Aktifkan profiling CPU dan heap")
	mysqlFlag := flag.String("mysql", "", "Custom MySQL DSN (kosongkan untuk mengambil dari .env)")
	flag.Parse()

	log.Println("[Eksperimen 3] Mulai Benchmark Database SQL INSERT vs Zero-Alloc Batch INSERT")

	// Load DSN & hubungkan ke database
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("[Config] Gagal memuat config: %v", err)
	}
	if *mysqlFlag != "" {
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName = "", "", "", "", ""
	}

	// Init MySQL menggunakan helper asli backend
	_, sqlDB, err := config.InitMysql(cfg)
	if err != nil {
		log.Fatalf("[MySQL] Gagal inisialisasi MySQL: %v. Pastikan database lokal menyala jika ingin menjalankan ini.", err)
	}
	defer sqlDB.Close()

	// Pastikan tabel kosong sebelum benchmark
	_, _ = sqlDB.Exec("TRUNCATE TABLE sensor_logs")

	if *pprofFlag {
		cpuPath := filepath.Join(*outputFlag, "cpu-db.prof")
		heapPath := filepath.Join(*outputFlag, "heap-db.prof")
		_ = os.MkdirAll(*outputFlag, 0755)
		profiler := profile.NewProfiler(cpuPath, heapPath)
		profiler.Start(cpuPath)
		defer profiler.Stop()
	}

	sensorRepo := repository.NewSensorRepository(sqlDB)

	// Uji Batch Size bervariasi
	batchSizes := []int{1, 10, 25, 50, 100, 250, 500}
	results := make(map[int]float64)
	allocResults := make(map[int]float64)

	// Buat data generator (deterministik) sebanyak 1000 baris
	dataset := generator.GenerateDeterministicDataset(1000, 45)

	// Warmup database
	reporter.WarmUp(2*time.Second, func() {
		_ = sensorRepo.BatchInsert(context.Background(), dataset[0:10])
		_, _ = sqlDB.Exec("TRUNCATE TABLE sensor_logs")
	})

	for _, bSize := range batchSizes {
		log.Printf("[Batch Size = %d] Menjalankan uji performa...\n", bSize)
		var rateRuns []float64
		var memRuns []float64

		for r := 1; r <= *countFlag; r++ {
			// Bersihkan tabel
			_, _ = sqlDB.Exec("TRUNCATE TABLE sensor_logs")

			tracker := metrics.NewResourceTracker()
			tracker.Start()

			start := time.Now()
			totalInserted := 0

			// Bagi dataset ke dalam batch-batch berukuran bSize
			for i := 0; i < len(dataset); i += bSize {
				end := i + bSize
				if end > len(dataset) {
					end = len(dataset)
				}
				batch := dataset[i:end]

				if bSize == 1 {
					// Single Insert Mode: Manual loop/insert per baris
					for _, item := range batch {
						_ = insertSingleRow(sqlDB, item)
					}
				} else {
					// Batch Insert Mode: Menggunakan Zero-Alloc BatchInsert bawaan prod
					_ = sensorRepo.BatchInsert(context.Background(), batch)
				}
				totalInserted += len(batch)
			}

			elapsed := time.Since(start)
			usage := tracker.Stop()

			insertsPerSec := float64(totalInserted) / elapsed.Seconds()
			rateRuns = append(rateRuns, insertsPerSec)
			memRuns = append(memRuns, usage.AllocatedMB)
		}

		results[bSize] = mean(rateRuns)
		allocResults[bSize] = mean(memRuns)
	}

	fmt.Println("\n=======================================================")
	fmt.Println("             EXPERIMENT 3: SQL INSERT PERFORMANCE       ")
	fmt.Println("=======================================================")
	fmt.Printf("Batch Size\tInserts / Second\tHeap Overhead (MB)\n")
	fmt.Println("-------------------------------------------------------")
	for _, bSize := range batchSizes {
		modeStr := fmt.Sprintf("%d", bSize)
		if bSize == 1 {
			modeStr = "1 (Single)"
		}
		fmt.Printf("%s\t\t%.2f rows/sec\t\t%.4f MB\n", modeStr, results[bSize], allocResults[bSize])
	}
	fmt.Println("=======================================================")

	// Ekspor ke file reporter
	params := map[string]interface{}{
		"count":        *countFlag,
		"dataset_size": 1000,
	}

	mResults := make(map[string]float64)
	for _, bSize := range batchSizes {
		mResults[fmt.Sprintf("batch_%d_inserts_sec", bSize)] = results[bSize]
		mResults[fmt.Sprintf("batch_%d_allocated_mb", bSize)] = allocResults[bSize]
	}

	exp := reporter.ExportData{
		Experiment: "db-insert",
		Params:     params,
		Metrics:    mResults,
	}

	// Simpan
	err = reporter.SaveResults(*outputFlag, exp, "", "MySQL 8.0+")
	if err != nil {
		log.Printf("[ERROR] Gagal menyimpan laporan eksperimen 3: %v\n", err)
	} else {
		log.Printf("[System] Hasil eksperimen 3 disimpan dengan sukses ke %s/\n", *outputFlag)
	}
}

func insertSingleRow(db *sql.DB, row model.SensorDataRow) error {
	query := "INSERT INTO sensor_logs (time, node_id, temperature, humidity) VALUES (?, ?, ?, ?)"
	_, err := db.Exec(query, row.Timestamp, row.NodeID, row.Temp, row.Humi)
	return err
}

func mean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}
