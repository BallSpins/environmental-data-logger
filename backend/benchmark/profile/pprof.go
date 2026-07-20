package profile

import (
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

type Profiler struct {
	cpuFile  *os.File
	heapPath string
}

func NewProfiler(cpuPath, heapPath string) *Profiler {
	return &Profiler{
		heapPath: heapPath,
	}
}

// Start merekam CPU profiling jika path ditentukan
func (p *Profiler) Start(cpuPath string) {
	if cpuPath == "" {
		return
	}

	var err error
	p.cpuFile, err = os.Create(cpuPath)
	if err != nil {
		log.Printf("[Profiler] Gagal membuat file CPU profile: %v\n", err)
		return
	}

	if err := pprof.StartCPUProfile(p.cpuFile); err != nil {
		log.Printf("[Profiler] Gagal memulai CPU profiling: %v\n", err)
		p.cpuFile.Close()
		p.cpuFile = nil
	} else {
		log.Printf("[Profiler] CPU Profiling dimulai -> %s\n", cpuPath)
	}
}

// Stop menghentikan perekaman CPU profile dan langsung menulis snapshot Heap Profile
func (p *Profiler) Stop() {
	if p.cpuFile != nil {
		pprof.StopCPUProfile()
		p.cpuFile.Close()
		log.Println("[Profiler] CPU Profiling dihentikan.")
		p.cpuFile = nil
	}

	if p.heapPath != "" {
		f, err := os.Create(p.heapPath)
		if err != nil {
			log.Printf("[Profiler] Gagal membuat file Heap profile: %v\n", err)
			return
		}
		defer f.Close()

		runtime.GC() // Paksakan GC sebelum merekam heap stat agar bersih
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Printf("[Profiler] Gagal menulis Heap profiling: %v\n", err)
		} else {
			log.Printf("[Profiler] Heap Profiling berhasil disimpan -> %s\n", p.heapPath)
		}
	}
}
