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
	"github.com/ballspins/environmental-data-logger/backend/benchmark/metrics"
	"github.com/ballspins/environmental-data-logger/backend/benchmark/profile"
	"github.com/ballspins/environmental-data-logger/backend/benchmark/reporter"
	"github.com/ballspins/environmental-data-logger/backend/internal/config"
	"github.com/ballspins/environmental-data-logger/backend/internal/repository"
)

func main() {
	nodesFlag := flag.Int("nodes", 100, "Jumlah virtual IoT nodes")
	durationFlag := flag.Duration("duration", 10*time.Second, "Durasi jalannya stress-test")
	outputFlag := flag.String("output", "results", "Direktori penyimpanan output")
	pprofFlag := flag.Bool("pprof", false, "Aktifkan profiling CPU dan heap")
	payloadFlag := flag.String("payload", "binary", "Format payload: binary atau json")
	publishRateFlag := flag.Float64("publish-rate", 0.05, "Frekuensi publish per node per detik (0.05 = 1 per 20s)")
	patternFlag := flag.String("pattern", "constant", "Traffic pattern: constant, burst, random")
	flag.Parse()

	log.Printf("[Eksperimen 5] Memulai Simulator Shadow Device (%d Nodes, Duration=%v, Payload=%s)\n", *nodesFlag, *durationFlag, *payloadFlag)

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("[Config] Gagal memuat config: %v", err)
	}

	// Koneksi MySQL & Redis
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
		cpuPath := filepath.Join(*outputFlag, "cpu-stress.prof")
		heapPath := filepath.Join(*outputFlag, "heap-stress.prof")
		_ = os.MkdirAll(*outputFlag, 0755)
		profiler := profile.NewProfiler(cpuPath, heapPath)
		profiler.Start(cpuPath)
		defer profiler.Stop()
	}

	cacheRepo := repository.NewCacheRepository(rdb)

	// Bersihkan Redis & MySQL logs untuk test yang deterministik
	ctx := context.Background()
	_ = rdb.Del(ctx, "queue:sensor_ingestion")
	_, _ = sqlDB.Exec("TRUNCATE TABLE sensor_logs")

	latencyTracker := metrics.NewLatencyTracker(10000)
	resourceTracker := metrics.NewResourceTracker()
	resourceTracker.Start()

	var mu sync.Mutex
	var publishCount float64
	var totalBytes float64

	// Callback yang dipanggil setiap kali node virtual mempublish data
	publishFunc := func(nodeID uint8) {
		startPub := time.Now()

		chunk := generator.GenerateRandomChunk(nodeID)
		var size float64

		if *payloadFlag == "binary" {
			var raw [48]byte
			generator.ChunkToBinary(chunk, raw[:])
			_ = cacheRepo.BufferSensorData(ctx, raw[:])
			size = 48.0
		} else {
			jsonBytes, _ := generator.ChunkToJSON(chunk)
			_ = cacheRepo.BufferSensorData(ctx, jsonBytes)
			size = float64(len(jsonBytes))
		}

		duration := time.Since(startPub)

		mu.Lock()
		publishCount++
		totalBytes += size
		latencyTracker.Record(duration)
		mu.Unlock()
	}

	// Inisialisasi scheduler & pool shadow devices
	scheduler := generator.NewNodeScheduler(*nodesFlag, *publishRateFlag, generator.TrafficPattern(*patternFlag), publishFunc)

	// Warm-up scheduler
	reporter.WarmUp(1*time.Second, func() {
		publishFunc(1)
	})

	_ = rdb.Del(ctx, "queue:sensor_ingestion")

	// Jalankan scheduler
	log.Printf("[Stress] Memulai emulasi scheduler dengan pattern: %s\n", *patternFlag)
	stressStart := time.Now()
	scheduler.Start()

	// Catat antrean per detik
	var periodicQueueLength []float64
	ticker := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				qLen, _ := rdb.LLen(ctx, "queue:sensor_ingestion").Result()
				periodicQueueLength = append(periodicQueueLength, float64(qLen))
			case <-time.After(*durationFlag):
				ticker.Stop()
				return
			}
		}
	}()

	time.Sleep(*durationFlag)
	scheduler.Stop()

	elapsedStress := time.Since(stressStart)
	systemUsage := resourceTracker.Stop()

	stats := latencyTracker.Calculate()

	// Ekstrak data antrean akhir
	finalQLen, _ := rdb.LLen(ctx, "queue:sensor_ingestion").Result()

	fmt.Println("\n=======================================================")
	fmt.Println("             EXPERIMENT 5: END-TO-END SCALABILITY      ")
	fmt.Println("=======================================================")
	fmt.Printf("Shadow Devices\t\t%d nodes\n", *nodesFlag)
	fmt.Printf("Total Duration\t\t%v\n", elapsedStress)
	fmt.Printf("Total MQTT Publish\t%.0f events\n", publishCount)
	fmt.Printf("Throughput Rate\t\t%.2f pub/sec\n", publishCount/elapsedStress.Seconds())
	fmt.Printf("Bandwidth Rate\t\t%.2f KB/s\n", (totalBytes/1024.0)/elapsedStress.Seconds())
	fmt.Printf("Mean Latency\t\t%v\n", stats.Mean)
	fmt.Printf("P50 Latency\t\t%v\n", stats.P50)
	fmt.Printf("P95 Latency\t\t%v\n", stats.P95)
	fmt.Printf("P99 Latency\t\t%v\n", stats.P99)
	fmt.Printf("Max Redis Queue Len\t%d msgs (Final: %d)\n", int(max(periodicQueueLength)), finalQLen)
	fmt.Printf("Go Heap Allocated\t%.2f MB\n", systemUsage.AllocatedMB)
	fmt.Println("=======================================================")

	// Ekspor ke reporter
	params := map[string]interface{}{
		"nodes":        *nodesFlag,
		"duration":     durationFlag.String(),
		"payload":      *payloadFlag,
		"publish_rate": *publishRateFlag,
		"pattern":      *patternFlag,
	}

	mResults := map[string]float64{
		"total_publishes":    publishCount,
		"throughput_pub_sec": publishCount / elapsedStress.Seconds(),
		"bandwidth_kb_sec":   (totalBytes / 1024.0) / elapsedStress.Seconds(),
		"mean_latency_ms":    float64(stats.Mean.Milliseconds()),
		"p50_latency_ms":     float64(stats.P50.Milliseconds()),
		"p95_latency_ms":     float64(stats.P95.Milliseconds()),
		"p99_latency_ms":     float64(stats.P99.Milliseconds()),
		"max_queue_len":      max(periodicQueueLength),
		"final_queue_len":    float64(finalQLen),
		"heap_allocated_mb":  systemUsage.AllocatedMB,
	}

	exp := reporter.ExportData{
		Experiment: "stress-test",
		Params:     params,
		Metrics:    mResults,
	}

	err = reporter.SaveResults(*outputFlag, exp, "Redis 6.2+", "MySQL 8.0+")
	if err != nil {
		log.Printf("[ERROR] Gagal menyimpan laporan eksperimen 5: %v\n", err)
	} else {
		log.Printf("[System] Hasil eksperimen 5 disimpan dengan sukses ke %s/\n", *outputFlag)
	}
}

func max(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	m := vals[0]
	for _, v := range vals {
		if v > m {
			m = v
		}
	}
	return m
}
