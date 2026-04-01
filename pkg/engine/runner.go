package engine

import (
	"bufio"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/loggling/loggling/pkg/model"
)

type LogItem struct {
	TimeStamp int64
	Data      []byte
}

type StreamRunner struct {
	pipeline atomic.Value
}

func NewStreamRunner(p *Pipeline) *StreamRunner {
	r := &StreamRunner{}
	r.pipeline.Store(p)
	return r
}

func (r *StreamRunner) SwapPipeline(p *Pipeline) {
	r.pipeline.Store(p)
}

func (r *StreamRunner) getPipeline() *Pipeline {
	return r.pipeline.Load().(*Pipeline)
}

func (r *StreamRunner) RunParallel(inputs []io.Reader, output io.Writer) error {
	numFiles := len(inputs)
	channels := make([]chan LogItem, numFiles)
	workerCounts := make([]int64, numFiles)

	for i := range numFiles {
		channels[i] = make(chan LogItem, 100)

		go r.worker(inputs[i], channels[i], &workerCounts[i])
	}

	stopTUI := make(chan bool)
	go r.renderTUI(workerCounts, stopTUI)
	err := r.mergeAndWrite(channels, output)

	close(stopTUI)

	return err
}

func (r *StreamRunner) mergeAndWrite(channels []chan LogItem, output io.Writer) error {
	numFiles := len(channels)
	heads := make([]*LogItem, numFiles)

	for i := range numFiles {
		item, ok := <-channels[i]

		if ok {
			heads[i] = &item
		} else {
			heads[i] = nil
		}
	}

	writer := bufio.NewWriter(output)
	defer writer.Flush()

	for {
		minIndex := -1
		var minTime int64 = 1<<63 - 1

		for i := range numFiles {
			if heads[i] != nil && heads[i].TimeStamp < minTime {
				minTime = heads[i].TimeStamp
				minIndex = i
			}
		}

		if minIndex == -1 {
			break
		}

		writer.Write(heads[minIndex].Data)
		writer.WriteByte('\n')

		nextItem, ok := <-channels[minIndex]

		if ok {
			heads[minIndex] = &nextItem
		} else {
			heads[minIndex] = nil
		}

	}

	return nil
}

func (r *StreamRunner) worker(input io.Reader, out chan<- LogItem, myCounter *int64) {

	scanner := bufio.NewScanner(input)
	var lastTime int64 = 0

	for scanner.Scan() {
		line := scanner.Bytes()
		p := r.getPipeline()
		result := p.Execute(line)

		if result != nil {
			currentTime := extractTimestamp()

			if currentTime == 0 {
				currentTime = lastTime
			} else {
				lastTime = currentTime
			}

			dataCopy := make([]byte, len(result.Data))
			copy(dataCopy, result.Data)

			out <- LogItem{
				TimeStamp: currentTime,
				Data:      dataCopy,
			}

			p.Release(result)
			atomic.AddInt64(myCounter, 1)
		}
	}
	close(out)
}

func extractTimestamp() int64 {
	return time.Now().UnixNano()
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

				fmt.Printf("\r [Loggling] Speed: %d logs/sec | Total: %d | Dropped: %d | Errored: %d",
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
		p := r.getPipeline()

		func() {
			defer func() {
				if r := recover(); r != nil {
					model.GlobalMetrics.AddErrorLine()
				}
			}()

			result := p.Execute(line)

			if result != nil {
				writer.Write(result.Data)
				writer.WriteByte('\n')
				p.Release(result)
			}
		}()
	}

	close(stop)
	fmt.Printf("\n All processing complete. Final processed: %d, dropped: %d\n",
		model.GlobalMetrics.ProcessedLines.Load(),
		model.GlobalMetrics.DroppedLines.Load())

	return scanner.Err()
}

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
