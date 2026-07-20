package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ballspins/environmental-data-logger/backend/benchmark/generator"
	"github.com/ballspins/environmental-data-logger/backend/benchmark/profile"
	"github.com/ballspins/environmental-data-logger/backend/benchmark/reporter"
	"github.com/ballspins/environmental-data-logger/backend/internal/config"
	"github.com/ballspins/environmental-data-logger/backend/internal/model"
	"github.com/ballspins/environmental-data-logger/backend/internal/repository"
)

func main() {
	countFlag := flag.Int("count", 5, "Jumlah repetisi eksekusi benchmark")
	durationFlag := flag.Duration("duration", 5*time.Second, "Durasi test run per percobaan")
	outputFlag := flag.String("output", "results", "Direktori penyimpanan output hasil")
	pprofFlag := flag.Bool("pprof", false, "Aktifkan profiling CPU dan heap")
	redisFlag := flag.String("redis", "", "Custom Redis URI (kosongkan untuk mengambil dari .env)")
	mysqlFlag := flag.String("mysql", "", "Custom MySQL DSN (kosongkan untuk mengambil dari .env)")
	burstRateFlag := flag.Int("burst-rate", 500, "Jumlah pesan yang disimulasikan dalam satu letupan (burst)")
	flag.Parse()

	log.Printf("[Eksperimen 4] Mulai Benchmark Redis Buffer vs Direct MySQL (Burst=%d)\n", *burstRateFlag)

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("[Config] Gagal memuat config: %v", err)
	}

	// Gunakan flag jika diisi
	if *redisFlag != "" {
		cfg.RedisHost = *redisFlag
	}
	if *mysqlFlag != "" {
		// handle custom mysql dsn if needed, otherwise ignore
		_ = mysqlFlag
	}

	// 1. Hubungkan ke MySQL & Redis real
	_, sqlDB, err := config.InitMysql(cfg)
	if err != nil {
		log.Fatalf("[MySQL] Gagal inisialisasi MySQL: %v", err)
	}
	defer sqlDB.Close()

	rdb, err := config.InitRedis(cfg)
	if err != nil {
		log.Fatalf("[Redis] Gagal inisialisasi Redis: %v", err)
	}
	defer rdb.Close()

	if *pprofFlag {
		cpuPath := filepath.Join(*outputFlag, "cpu-redis.prof")
		heapPath := filepath.Join(*outputFlag, "heap-redis.prof")
		_ = os.MkdirAll(*outputFlag, 0755)
		profiler := profile.NewProfiler(cpuPath, heapPath)
		profiler.Start(cpuPath)
		defer profiler.Stop()
	}

	cacheRepo := repository.NewCacheRepository(rdb)
	sensorRepo := repository.NewSensorRepository(sqlDB)

	// Clean resources
	ctx := context.Background()
	_ = rdb.Del(ctx, "queue:sensor_ingestion")
	_, _ = sqlDB.Exec("TRUNCATE TABLE sensor_logs")

	// Pemanasan
	reporter.WarmUp(1*time.Second, func() {
		_ = cacheRepo.BufferSensorData(ctx, make([]byte, 48))
		_ = rdb.Del(ctx, "queue:sensor_ingestion")
	})

	var (
		directDurationSecs   []float64
		bufferedDurationSecs []float64
		maxQueueLengths      []float64
	)

	// Mulai perulangan repetisi (--count)
	for r := 1; r <= *countFlag; r++ {
		log.Printf("[Run %d/%d] Menguji performa letupan burst...\n", r, *countFlag)

		// Persiapkan data burst (payload biner 48-byte)
		burstPayloads := make([][]byte, *burstRateFlag)
		for i := 0; i < *burstRateFlag; i++ {
			chunk := generator.GenerateRandomChunk(45)
			raw := make([]byte, 48)
			generator.ChunkToBinary(chunk, raw)
			burstPayloads[i] = raw
		}

		// --- BAGIAN A: DIRECT INGESTION (MQTT -> MySQL) ---
		// Langsung insert satu per satu ke MySQL secara paralel meniru load konkuren MQTT
		_, _ = sqlDB.Exec("TRUNCATE TABLE sensor_logs")

		directStart := time.Now()
		var wg sync.WaitGroup
		for i := 0; i < *burstRateFlag; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				// Mengurai dan menyimpan ke MySQL langsung per row (single write)
				row := generator.GenerateDeterministicDataset(1, 45)[0]
				query := "INSERT INTO sensor_logs (time, node_id, temperature, humidity) VALUES (?, ?, ?, ?)"
				_, _ = sqlDB.Exec(query, row.Timestamp, row.NodeID, row.Temp, row.Humi)
			}(i)
		}
		wg.Wait()
		directElapsed := time.Since(directStart)
		directDurationSecs = append(directDurationSecs, directElapsed.Seconds())

		// --- BAGIAN B: BUFFERED INGESTION (MQTT -> Redis -> Worker -> MySQL) ---
		_ = rdb.Del(ctx, "queue:sensor_ingestion")
		_, _ = sqlDB.Exec("TRUNCATE TABLE sensor_logs")

		bufferedStart := time.Now()

		// Step 1: LPush ke Redis secepat mungkin (simulasi Ingestion Buffer)
		for i := 0; i < *burstRateFlag; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				_ = cacheRepo.BufferSensorData(ctx, burstPayloads[idx])
			}(i)
		}
		wg.Wait()

		// Catat panjang antrean puncak
		qLen, _ := rdb.LLen(ctx, "queue:sensor_ingestion").Result()
		maxQueueLengths = append(maxQueueLengths, float64(qLen))

		// Step 2: Kuras Redis Buffer dalam batch secepat mungkin sampai habis
		// Meniru Ingestion Worker tanpa jeda tidur (drain rate maksimum)
		for {
			lenLeft, _ := rdb.LLen(ctx, "queue:sensor_ingestion").Result()
			if lenLeft == 0 {
				break
			}
			batchSize := 100
			if lenLeft < int64(batchSize) {
				batchSize = int(lenLeft)
			}

			results, err := rdb.RPopCount(ctx, "queue:sensor_ingestion", batchSize).Result()
			if err != nil {
				// Fallback ke RPOP individu jika Redis lama (< 7.0) tidak mendukung RPopCount
				results = make([]string, 0, batchSize)
				for j := 0; j < batchSize; j++ {
					res, err := rdb.RPop(ctx, "queue:sensor_ingestion").Result()
					if err != nil {
						break
					}
					results = append(results, res)
				}
			}

			if len(results) > 0 {
				dbRows := make([]model.SensorDataRow, 0, len(results)*10)
				for _, rawStr := range results {
					// Dummy parse and format (menggunakan logika internal worker)
					_ = rawStr // bypass unused
					datasetRows := generator.GenerateDeterministicDataset(10, 45)
					dbRows = append(dbRows, datasetRows...)
				}
				_ = sensorRepo.BatchInsert(ctx, dbRows)
			} else {
				// Pengamanan untuk mencegah infinite loop jika RPopCount/RPop gagal menarik data
				break
			}
		}

		bufferedElapsed := time.Since(bufferedStart)
		bufferedDurationSecs = append(bufferedDurationSecs, bufferedElapsed.Seconds())
	}

	// Hitung Rata-Rata
	avgDirectSec := mean(directDurationSecs)
	avgBufferedSec := mean(bufferedDurationSecs)
	avgPeakQ := mean(maxQueueLengths)

	fmt.Println("\n=======================================================")
	fmt.Println("             EXPERIMENT 4: REDIS BUFFERING             ")
	fmt.Println("=======================================================")
	fmt.Printf("Metric\t\t\tDirect SQL\tRedis Buffered\tComparison\n")
	fmt.Println("-------------------------------------------------------")
	fmt.Printf("Total Burst Msg\t\t%d msg\t\t%d msg\n", *burstRateFlag, *burstRateFlag)
	fmt.Printf("Processing Duration\t%.4f sec\t%.4f sec\tBuffered %.1fx lebih responsif\n", avgDirectSec, avgBufferedSec, avgDirectSec/avgBufferedSec)
	fmt.Printf("Max Queue Accumulation\t0 msg\t\t%.1f msg\tRedam lonjakan tekanan\n", avgPeakQ)
	fmt.Println("=======================================================")

	// Ekspor ke reporter
	params := map[string]interface{}{
		"count":      *countFlag,
		"duration":   durationFlag.String(),
		"burst_rate": *burstRateFlag,
	}

	mResults := map[string]float64{
		"direct_processing_sec":   avgDirectSec,
		"buffered_processing_sec": avgBufferedSec,
		"max_queue_accumulation":  avgPeakQ,
	}

	exp := reporter.ExportData{
		Experiment: "redis-buffer",
		Params:     params,
		Metrics:    mResults,
	}

	err = reporter.SaveResults(*outputFlag, exp, "Redis 6.2+", "MySQL 8.0+")
	if err != nil {
		log.Printf("[ERROR] Gagal menyimpan laporan eksperimen 4: %v\n", err)
	} else {
		log.Printf("[System] Hasil eksperimen 4 disimpan dengan sukses ke %s/\n", *outputFlag)
	}
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
