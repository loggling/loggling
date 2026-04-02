package engine

import (
	"fmt"
	"sync/atomic"
	"time"
)

func (r *StreamRunner) renderTUI(counts []int64, stop <-chan bool) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	numWorkers := len(counts)
	lastCounts := make([]int64, numWorkers)
	firstDraw := true

	for {
		select {
		case <-ticker.C:
			if !firstDraw {
				fmt.Printf("\033[%dA", numWorkers+1)
			}

			firstDraw = false

			var totalTPS int64
			var totalLines int64

			for i := range numWorkers {
				current := atomic.LoadInt64(&counts[i])
				tps := current - lastCounts[i]

				lastCounts[i] = current

				totalTPS += tps
				totalLines += current
				fmt.Printf("Worker %02d | Speed: %8d logs/s | Total: %10d \033[K\n", i+1, tps, current)
			}
			fmt.Printf("[TOTAL]  | Speed: %8d logs/s | Total: %10d \033[K\n", totalTPS, totalLines)

		case <-stop:
			fmt.Println("success")
			return
		}
	}
}
