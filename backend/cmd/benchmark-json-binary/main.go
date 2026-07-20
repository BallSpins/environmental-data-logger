package main

import (
	"encoding/json"
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
	"github.com/ballspins/environmental-data-logger/backend/internal/model"
	"github.com/ballspins/environmental-data-logger/backend/internal/service"
)

func main() {
	countFlag := flag.Int("count", 5, "Jumlah repetisi eksekusi benchmark")
	durationFlag := flag.Duration("duration", 3*time.Second, "Durasi pengerjaan per iterasi")
	outputFlag := flag.String("output", "results", "Direktori penyimpanan output hasil")
	pprofFlag := flag.Bool("pprof", false, "Aktifkan profiling CPU dan heap")
	flag.Parse()

	log.Printf("[Eksperimen 1] Mulai Benchmark JSON vs Binary (Count=%d, Duration=%v)\n", *countFlag, *durationFlag)

	// Inisialisasi Profiling jika aktif
	var profiler *profile.Profiler
	if *pprofFlag {
		cpuPath := filepath.Join(*outputFlag, "cpu-json-binary.prof")
		heapPath := filepath.Join(*outputFlag, "heap-json-binary.prof")
		_ = os.MkdirAll(*outputFlag, 0755)
		profiler = profile.NewProfiler(cpuPath, heapPath)
		profiler.Start(cpuPath)
		defer profiler.Stop()
	}

	// Warm-up Parser Service
	parser := service.NewParserService()
	chunkSample := generator.GenerateRandomChunk(45)
	var binaryBuf [48]byte
	reporter.WarmUp(2*time.Second, func() {
		generator.ChunkToBinary(chunkSample, binaryBuf[:])
		var dummy model.ChunkPayload
		_ = parser.ParseBinary(binaryBuf[:], &dummy)
	})

	// Penampung Hasil
	var (
		binarySerialNs    []float64
		binaryDeserialNs  []float64
		binaryAllocMB     []float64
		binaryPayloadSize []float64
		binaryGzipSize    []float64

		jsonSerialNs    []float64
		jsonDeserialNs  []float64
		jsonAllocMB     []float64
		jsonPayloadSize []float64
		jsonGzipSize    []float64
	)

	// Lakukan perulangan untuk repetisi (--count)
	for r := 1; r <= *countFlag; r++ {
		log.Printf("[Run %d/%d] Mengeksekusi pengukuran...\n", r, *countFlag)

		// --- 1. BINARY EXPERIMENT ---
		binTracker := metrics.NewResourceTracker()
		binTracker.Start()

		binStart := time.Now()
		var iterations int
		var binSize, binGzip float64

		for time.Since(binStart) < *durationFlag {
			iterations++
			chunk := generator.GenerateRandomChunk(45)

			// Serialisasi ke Biner 48-byte
			var raw [48]byte
			generator.ChunkToBinary(chunk, raw[:])
			binSize = float64(len(raw))

			if r == 1 && iterations == 1 {
				gz, _ := generator.CompressGzip(raw[:])
				binGzip = float64(len(gz))
			}

			// Deserialisasi kembali
			var decoded model.ChunkPayload
			_ = parser.ParseBinary(raw[:], &decoded)
		}

		binElapsed := time.Since(binStart)
		binUsage := binTracker.Stop()

		avgBinSerialNs := float64(binElapsed.Nanoseconds()) / float64(iterations) / 2.0 // estimasi serialisasi + deserialisasi berimbang
		binarySerialNs = append(binarySerialNs, avgBinSerialNs)
		binaryDeserialNs = append(binaryDeserialNs, avgBinSerialNs)
		binaryAllocMB = append(binaryAllocMB, binUsage.AllocatedMB)
		binaryPayloadSize = append(binaryPayloadSize, binSize)
		binaryGzipSize = append(binaryGzipSize, binGzip)

		// --- 2. JSON EXPERIMENT ---
		jsonTracker := metrics.NewResourceTracker()
		jsonTracker.Start()

		jsonStart := time.Now()
		jsonIterations := 0
		var jsonSize, jsonGzip float64

		for time.Since(jsonStart) < *durationFlag {
			jsonIterations++
			chunk := generator.GenerateRandomChunk(45)

			// Serialisasi ke JSON
			jsonBytes, _ := generator.ChunkToJSON(chunk)
			jsonSize = float64(len(jsonBytes))

			if r == 1 && jsonIterations == 1 {
				gz, _ := generator.CompressGzip(jsonBytes)
				jsonGzip = float64(len(gz))
			}

			// Deserialisasi kembali
			var decoded generator.JSONPayload
			_ = json.Unmarshal(jsonBytes, &decoded)
		}

		jsonElapsed := time.Since(jsonStart)
		jsonUsage := jsonTracker.Stop()

		avgJsonSerialNs := float64(jsonElapsed.Nanoseconds()) / float64(jsonIterations) / 2.0
		jsonSerialNs = append(jsonSerialNs, avgJsonSerialNs)
		jsonDeserialNs = append(jsonDeserialNs, avgJsonSerialNs)
		jsonAllocMB = append(jsonAllocMB, jsonUsage.AllocatedMB)
		jsonPayloadSize = append(jsonPayloadSize, jsonSize)
		jsonGzipSize = append(jsonGzipSize, jsonGzip)
	}

	// Kalkulasi Statistik Akhir
	binSerialMean := mean(binarySerialNs)
	binAllocMean := mean(binaryAllocMB)
	jsonSerialMean := mean(jsonSerialNs)
	jsonAllocMean := mean(jsonAllocMB)

	fmt.Println("\n=======================================================")
	fmt.Println("             EXPERIMENT 1: JSON vs BINARY              ")
	fmt.Println("=======================================================")
	fmt.Printf("Metric\t\t\tBinary\t\tJSON\t\tComparison\n")
	fmt.Println("-------------------------------------------------------")
	fmt.Printf("Payload Size (B)\t%.1f B\t\t%.1f B\t\tBinary %.1fx lebih hemat\n", binaryPayloadSize[0], jsonPayloadSize[0], jsonPayloadSize[0]/binaryPayloadSize[0])
	fmt.Printf("Gzip Compressed (B)\t%.1f B\t\t%.1f B\t\tBinary %.1fx lebih hemat\n", binaryGzipSize[0], jsonGzipSize[0], jsonGzipSize[0]/binaryGzipSize[0])
	fmt.Printf("Serialization Time\t%.2f ns/op\t%.2f ns/op\tBinary %.1fx lebih cepat\n", binSerialMean, jsonSerialMean, jsonSerialMean/binSerialMean)
	fmt.Printf("Heap Allocated\t\t%.4f MB/op\t%.4f MB/op\tBinary %.1fx lebih bersih\n", binAllocMean, jsonAllocMean, (jsonAllocMean+1e-9)/(binAllocMean+1e-9))
	fmt.Println("=======================================================")

	// Ekspor ke file reporter
	params := map[string]interface{}{
		"duration":   durationFlag.String(),
		"count":      *countFlag,
		"chunk_size": 10,
	}

	mResults := map[string]float64{
		"binary_payload_size_bytes": binaryPayloadSize[0],
		"binary_gzip_size_bytes":    binaryGzipSize[0],
		"binary_serial_ns_op":       binSerialMean,
		"binary_heap_allocated_mb":  binAllocMean,
		"json_payload_size_bytes":   jsonPayloadSize[0],
		"json_gzip_size_bytes":      jsonGzipSize[0],
		"json_serial_ns_op":         jsonSerialMean,
		"json_heap_allocated_mb":    jsonAllocMean,
	}

	exp := reporter.ExportData{
		Experiment: "json-vs-binary",
		Params:     params,
		Metrics:    mResults,
	}

	err := reporter.SaveResults(*outputFlag, exp, "", "")
	if err != nil {
		log.Printf("[ERROR] Gagal menyimpan laporan eksperimen 1: %v\n", err)
	} else {
		log.Printf("[System] Hasil eksperimen 1 disimpan dengan sukses ke %s/\n", *outputFlag)
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
