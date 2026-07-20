package generator

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

type TrafficPattern string

const (
	PatternConstant TrafficPattern = "constant"
	PatternBurst    TrafficPattern = "burst"
	PatternRandom   TrafficPattern = "random"
)

// NodeScheduler mengatur pengiriman data telemetri dari node virtual dengan berbagai pattern
type NodeScheduler struct {
	NumNodes    int
	PublishRate float64 // rata-rata publish per detik per node
	Pattern     TrafficPattern
	PublishFunc func(nodeID uint8)
	workerWg    sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewNodeScheduler(numNodes int, publishRate float64, pattern TrafficPattern, publishFunc func(nodeID uint8)) *NodeScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &NodeScheduler{
		NumNodes:    numNodes,
		PublishRate: publishRate,
		Pattern:     pattern,
		PublishFunc: publishFunc,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (s *NodeScheduler) Start() {
	if s.NumNodes <= 0 {
		return
	}

	// Interval dasar per node (misal publishRate = 0.05 -> 1 publish tiap 20 detik)
	baseInterval := time.Duration(float64(time.Second) / s.PublishRate)

	// Gunakan Worker Pool berukuran proporsional untuk meminimalkan beban idle goroutines
	numWorkers := s.NumNodes
	if numWorkers > 1000 {
		numWorkers = 1000 // Batasi worker aktif maksimal 1000 untuk efisiensi pool
	}

	taskChan := make(chan uint8, s.NumNodes*2)

	// Spawning workers
	for i := 0; i < numWorkers; i++ {
		s.workerWg.Add(1)
		go func() {
			defer s.workerWg.Done()
			for {
				select {
				case <-s.ctx.Done():
					return
				case nodeID, ok := <-taskChan:
					if !ok {
						return
					}
					s.PublishFunc(nodeID)
				}
			}
		}()
	}

	// Dispatcher loop berdasarkan pattern
	go func() {
		defer close(taskChan)
		ticker := time.NewTicker(baseInterval / time.Duration(s.NumNodes))
		defer ticker.Stop()

		var burstActive bool
		var burstCount int

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				switch s.Pattern {
				case PatternConstant:
					// Distribusikan pemanggilan nodeID secara bergantian secara konstan
					for i := 1; i <= s.NumNodes; i++ {
						select {
						case taskChan <- uint8((i-1)%255 + 1):
						case <-s.ctx.Done():
							return
						default:
							// Skip jika channel penuh
						}
					}

				case PatternRandom:
					// Pilih node acak dan kirimkan dengan sedikit variasi waktu acak
					for i := 0; i < s.NumNodes; i++ {
						randomNode := uint8(rand.Intn(s.NumNodes)%255 + 1)
						// Tambahkan jitter acak
						time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
						select {
						case taskChan <- randomNode:
						case <-s.ctx.Done():
							return
						default:
						}
					}

				case PatternBurst:
					// Simulasi burst: kirim banyak data sekaligus, lalu diam beberapa saat
					if !burstActive {
						burstActive = true
						burstCount = s.NumNodes * 3 // Kirim 3x lipat load sekaligus
						for j := 0; j < burstCount; j++ {
							nodeID := uint8(rand.Intn(s.NumNodes)%255 + 1)
							select {
							case taskChan <- nodeID:
							case <-s.ctx.Done():
								return
							default:
							}
						}
						// Tunggu burst selesai sebelum kembali idle
						time.Sleep(baseInterval * 3)
						burstActive = false
					}
				}
			}
		}
	}()
}

func (s *NodeScheduler) Stop() {
	s.cancel()
	s.workerWg.Wait()
}
