// Package model defines the core data structures and metrics used by the engine.
// metrics.go provides thread-safe counters for monitoring log processing engine performance.
package model

import "sync/atomic"

type Metrics struct {
	ProcessedLines atomic.Uint64
	ProcessedBytes atomic.Uint64
	DroppedLines   atomic.Uint64
	ErrorLines     atomic.Uint64
	TotalLatency   atomic.Uint64
}

var GlobalMetrics = &Metrics{}

func (m *Metrics) AddProcessedLine(size int) {
	m.ProcessedLines.Add(1)
	m.ProcessedBytes.Add((uint64(size)))
}

func (m *Metrics) AddDroppedLine() {
	m.DroppedLines.Add(1)
}

func (m *Metrics) AddErrorLine() {
	m.ErrorLines.Add(1)
}

func (m *Metrics) AddLatency(nanos int64) {
	m.TotalLatency.Add(uint64(nanos))
}
