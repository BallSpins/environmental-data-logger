package main

import (
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
)

func main() {
	countFlag := flag.Int("count", 5, "Jumlah repetisi eksekusi benchmark")
	durationFlag := flag.Duration("duration", 3*time.Second, "Durasi pengerjaan per iterasi")
	outputFlag := flag.String("output", "results", "Direktori penyimpanan output hasil")
	pprofFlag := flag.Bool("pprof", false, "Aktifkan profiling CPU dan heap")
	flag.Parse()

	log.Printf("[Eksperimen 2] Mulai Benchmark Chunk Aggregation (Count=%d, Duration=%v)\n", *countFlag, *durationFlag)

	if *pprofFlag {
		cpuPath := filepath.Join(*outputFlag, "cpu-chunk.prof")
		heapPath := filepath.Join(*outputFlag, "heap-chunk.prof")
		_ = os.MkdirAll(*outputFlag, 0755)
		profiler := profile.NewProfiler(cpuPath, heapPath)
		profiler.Start(cpuPath)
		defer profiler.Stop()
	}

	sizes := []int{1, 10, 30, 60}
	results := make(map[int]float64)

	// Warm-up
	reporter.WarmUp(1*time.Second, func() {
		_ = generator.GenerateRandomChunk(45)
	})

	for _, size := range sizes {
		log.Printf("[Chunk Size = %d] Mengukur performa transmisi biner...\n", size)

		var iterationResults []float64

		for r := 1; r <= *countFlag; r++ {
			tracker := metrics.NewResourceTracker()
			tracker.Start()

			start := time.Now()
			iterations := 0

			for time.Since(start) < *durationFlag {
				iterations++
				// Buat payload dummy biner sebanding dengan chunk size
				_ = make([]byte, 8+size*4)
			}

			elapsed := time.Since(start)
			tracker.Stop()

			opsPerSec := float64(iterations) / elapsed.Seconds()
			iterationResults = append(iterationResults, opsPerSec)
		}

		avgOps := mean(iterationResults)
		results[size] = avgOps
	}

	fmt.Println("\n=======================================================")
	fmt.Println("             EXPERIMENT 2: CHUNK AGGREGATION           ")
	fmt.Println("=======================================================")
	fmt.Printf("Chunk Size\tEst. Payload Size (B)\tThroughput (ops/sec)\n")
	fmt.Println("-------------------------------------------------------")
	for _, size := range sizes {
		payloadSize := 8 + size*4
		fmt.Printf("%d logs\t\t%d B\t\t\t%.2f ops/sec\n", size, payloadSize, results[size])
	}
	fmt.Println("=======================================================")

	// Ekspor ke reporter
	params := map[string]interface{}{
		"duration": durationFlag.String(),
		"count":    *countFlag,
	}

	mResults := make(map[string]float64)
	for _, size := range sizes {
		mResults[fmt.Sprintf("chunk_%d_ops_sec", size)] = results[size]
		mResults[fmt.Sprintf("chunk_%d_payload_size_bytes", size)] = float64(8 + size*4)
	}

	exp := reporter.ExportData{
		Experiment: "chunk-aggregation",
		Params:     params,
		Metrics:    mResults,
	}

	err := reporter.SaveResults(*outputFlag, exp, "", "")
	if err != nil {
		log.Printf("[ERROR] Gagal menyimpan laporan eksperimen 2: %v\n", err)
	} else {
		log.Printf("[System] Hasil eksperimen 2 disimpan dengan sukses ke %s/\n", *outputFlag)
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
