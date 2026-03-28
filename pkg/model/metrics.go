package model

import "sync/atomic"

type Metrics struct {
	ProcessedLines atomic.Uint64 // 성공적으로 처리된 라인 수
	ProcessedBytes atomic.Uint64 // 처리된 총 바이트 수
	DroppedLines   atomic.Uint64 // 필터에 의해 드랍된 라인 수
	ErrorLines     atomic.Uint64 // 에러(패닉 등)가 발생한 라인 수
	TotalLatency   atomic.Uint64 // 나노초 단위의 총 지연 시간 (합산용)
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
