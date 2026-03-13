package engine

import (
	"bufio"
	"fmt"
	"io"
	"time"

	"github.com/loggling/loggling/pkg/model"
)

type StreamRunner struct {
	Pipeline *Pipeline
}

func (r *StreamRunner) Run(input io.Reader, output io.Writer) error {

	stop := make(chan bool)

	go func() {
		ticker := time.NewTicker(time.Second)

		defer ticker.Stop()
		var lastCount uint64

		for {
			select {
			case <-ticker.C:
				current := model.GlobalMetrics.ProcessedLines.Load()
				dropped := model.GlobalMetrics.DroppedLines.Load()
				errored := model.GlobalMetrics.ErrorLines.Load()
				tps := current - lastCount
				lastCount = current

				fmt.Printf("\r📊 [Loggling] Speed: %d logs/sec | Total: %d | Dropped: %d | Errored: %d",
					tps, current, dropped, errored)

			case <-stop:
				return
			}
		}
	}()

	scanner := bufio.NewScanner(input)
	writer := bufio.NewWriter(output)

	defer writer.Flush()

	for scanner.Scan() {
		line := scanner.Bytes()

		func() {
			defer func() {
				if r := recover(); r != nil {
					model.GlobalMetrics.AddErrorLine()
				}
			}()

			result := r.Pipeline.Execute(line)

			if result != nil {
				writer.Write(result.Data)
				writer.WriteByte('\n')
				r.Pipeline.Release(result)
			}
		}()
	}

	close(stop)
	fmt.Printf("\n✅ All processing complete. Final processed: %d, dropped: %d\n",
		model.GlobalMetrics.ProcessedLines.Load(),
		model.GlobalMetrics.DroppedLines.Load())

	return scanner.Err()
}
