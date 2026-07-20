package reporter

import (
	"fmt"
	"time"
)

// WarmUp mengeksekusi fungsi target berulang kali selama durasi yang ditentukan agar pool & runtime optimal
func WarmUp(duration time.Duration, task func()) {
	fmt.Printf("[Warm-up] Memulai pemanasan sistem selama %v...\n", duration)
	start := time.Now()
	for time.Since(start) < duration {
		task()
	}
	fmt.Println("[Warm-up] Pemanasan selesai. Keadaan runtime dan pool sekarang optimal.")
}
