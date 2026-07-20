package metrics

import (
	"runtime"
	"runtime/metrics"
	"time"
)

// ResourceTracker mengumpulkan performa CPU & memory murni menggunakan Go stdlib
type ResourceTracker struct {
	startMemStats runtime.MemStats
	endMemStats   runtime.MemStats
	cpuMetrics    []metrics.Sample
}

func NewResourceTracker() *ResourceTracker {
	// Menyiapkan deskriptor metric sampel untuk Go CPU (CPU time dalam detik user+system)
	samples := make([]metrics.Sample, 2)
	samples[0].Name = "/sched/latencies:seconds" // Scheduler latency
	samples[1].Name = "/gc/pauses:seconds"       // GC pause times
	return &ResourceTracker{
		cpuMetrics: samples,
	}
}

// Start memicu snapshot awal dari memori & runtime metrics
func (rt *ResourceTracker) Start() {
	runtime.GC() // Paksakan GC sebelum test agar baseline bersih
	runtime.ReadMemStats(&rt.startMemStats)
}

// SystemUsage merepresentasikan metrik internal Go runtime pasca benchmark
type SystemUsage struct {
	AllocatedMB  float64       // Jumlah memori baru yang dialokasikan di heap selama benchmark (MB)
	HeapObjects  uint64        // Selisih jumlah heap objects
	Mallocs      uint64        // Jumlah pemanggilan alokasi memori
	Frees        uint64        // Jumlah objek yang dibebaskan
	NumGC        uint32        // Berapa kali GC berjalan selama benchmark
	TotalGCPause time.Duration // Total jeda waktu akibat GC
	AveragePause time.Duration // Rata-rata waktu GC pause
}

// Stop menghitung selisih resource memori pasca pengujian
func (rt *ResourceTracker) Stop() SystemUsage {
	runtime.ReadMemStats(&rt.endMemStats)

	// Hitung alokasi memori heap bersih (end - start)
	allocDiffBytes := int64(rt.endMemStats.TotalAlloc) - int64(rt.startMemStats.TotalAlloc)
	if allocDiffBytes < 0 {
		allocDiffBytes = 0
	}
	allocatedMB := float64(allocDiffBytes) / (1024 * 1024)

	// Total GC pause diff
	gcPauseDiff := rt.endMemStats.PauseTotalNs - rt.startMemStats.PauseTotalNs
	numGCDiff := rt.endMemStats.NumGC - rt.startMemStats.NumGC

	var avgPause time.Duration
	if numGCDiff > 0 {
		avgPause = time.Duration(gcPauseDiff / uint64(numGCDiff))
	}

	return SystemUsage{
		AllocatedMB:  allocatedMB,
		HeapObjects:  rt.endMemStats.HeapObjects - rt.startMemStats.HeapObjects,
		Mallocs:      rt.endMemStats.Mallocs - rt.startMemStats.Mallocs,
		Frees:        rt.endMemStats.Frees - rt.startMemStats.Frees,
		NumGC:        numGCDiff,
		TotalGCPause: time.Duration(gcPauseDiff),
		AveragePause: avgPause,
	}
}
