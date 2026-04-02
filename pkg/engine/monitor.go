// Package engine provides the core log processing mechanics.
// monitor.go separates the Terminal UI (TUI) metrics rendering,
// displaying real-time processing speeds, dropped lines, and total parsed payloads.
package engine

import (
	"fmt"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/loggling/loggling/pkg/model/logger"
)

const tuiRefreshRate = time.Second

func (r *StreamRunner) renderTUI(counts []int64, names []string, stop <-chan bool) {
	ticker := time.NewTicker(tuiRefreshRate)
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
				fileName := filepath.Base(names[i])
				logger.Raw("Speed:", fmt.Sprintf("%8d", tps), "logs/s | Total:", fmt.Sprintf("%10d", current), ">", fileName, "\033[K")
			}
			logger.Raw("Speed:", fmt.Sprintf("%8d", totalTPS), "logs/s | Total:", fmt.Sprintf("%10d", totalLines), "> TOTAL", "\033[K")

		case <-stop:
			return
		}
	}
}
