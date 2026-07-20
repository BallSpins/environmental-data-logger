package metrics

import (
	"math"
	"sort"
	"time"
)

// LatencyTracker menyimpan durasi dalam nanodetik untuk mengukur distribusi latency
type LatencyTracker struct {
	durations []time.Duration
}

func NewLatencyTracker(capacity int) *LatencyTracker {
	return &LatencyTracker{
		durations: make([]time.Duration, 0, capacity),
	}
}

func (l *LatencyTracker) Record(d time.Duration) {
	l.durations = append(l.durations, d)
}

func (l *LatencyTracker) Reset() {
	l.durations = l.durations[:0]
}

// Stats merepresentasikan metrik statistik lengkap hasil benchmark
type Stats struct {
	Min    time.Duration
	Max    time.Duration
	Mean   time.Duration
	Median time.Duration
	P50    time.Duration
	P90    time.Duration
	P95    time.Duration
	P99    time.Duration
	StdDev time.Duration
	Count  int
}

// Calculate menghitung nilai statistik dari data latency terkumpul
func (l *LatencyTracker) Calculate() Stats {
	count := len(l.durations)
	if count == 0 {
		return Stats{}
	}

	// Duplikasi slice untuk disorting agar data asli tidak terganggu order-nya
	sorted := make([]time.Duration, count)
	copy(sorted, l.durations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	min := sorted[0]
	max := sorted[count-1]

	var sum int64
	for _, d := range sorted {
		sum += d.Nanoseconds()
	}
	meanNs := sum / int64(count)

	// Standar Deviasi
	var sumSqDiff float64
	meanFloat := float64(meanNs)
	for _, d := range sorted {
		diff := float64(d.Nanoseconds()) - meanFloat
		sumSqDiff += diff * diff
	}
	stdDevNs := math.Sqrt(sumSqDiff / float64(count))

	// Percentiles
	p50 := sorted[int(float64(count)*0.50)]
	p90 := sorted[int(float64(count)*0.90)]
	p95 := sorted[int(float64(count)*0.95)]
	p99 := sorted[int(float64(count)*0.99)]
	if count > 0 {
		// handle out of bound rounding
		p50 = sorted[clamp(int(float64(count)*0.50), 0, count-1)]
		p90 = sorted[clamp(int(float64(count)*0.90), 0, count-1)]
		p95 = sorted[clamp(int(float64(count)*0.95), 0, count-1)]
		p99 = sorted[clamp(int(float64(count)*0.99), 0, count-1)]
	}

	return Stats{
		Min:    min,
		Max:    max,
		Mean:   time.Duration(meanNs),
		Median: p50,
		P50:    p50,
		P90:    p90,
		P95:    p95,
		P99:    p99,
		StdDev: time.Duration(int64(stdDevNs)),
		Count:  count,
	}
}

func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
