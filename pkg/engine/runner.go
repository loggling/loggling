package engine

import (
	"bufio"
	"fmt"
	"io"
	"sync"
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
	dlq      *RotatableWriter
}

func NewStreamRunner(p *Pipeline, dlqWriter *RotatableWriter) *StreamRunner {
	r := &StreamRunner{
		dlq: dlqWriter,
	}
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
	mergedChan := make(chan *model.LogPayload, 5000)
	workerCounts := make([]int64, numFiles)
	var wg sync.WaitGroup

	for i := range numFiles {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			r.worker(inputs[idx], mergedChan, &workerCounts[idx])
		}(i)
	}

	stopTUI := make(chan bool)
	go r.renderTUI(workerCounts, stopTUI)

	// 파이프(채널)를 닫아주는 안전장치
	go func() {
		wg.Wait()
		close(mergedChan)
	}()

	writer := bufio.NewWriter(output)
	// 병합 쓰기 (Fan-in) 로직
	for result := range mergedChan {
		writer.Write(result.Data)
		writer.WriteByte('\n')
		p := r.getPipeline()
		p.Release(result)
	}
	writer.Flush()

	close(stopTUI)
	return nil
}

func (r *StreamRunner) worker(input io.Reader, out chan<- *model.LogPayload, myCounter *int64) {
	scanner := bufio.NewScanner(input)

	for scanner.Scan() {
		line := scanner.Bytes()
		p := r.getPipeline()
		result, err := p.Execute(line)

		if err != nil {
			if r.dlq != nil {
				r.dlq.Write(line)
				r.dlq.Write([]byte{'\n'})
			}
			continue
		}

		if result != nil {
			out <- result
			atomic.AddInt64(myCounter, 1)
		}
	}
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

			result, err := p.Execute(line)

			if err != nil {
				if r.dlq != nil {
					r.dlq.Write(line)
					r.dlq.Write([]byte{'\n'})
				}
				return
			}
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
