// Package model defines the core data structures and metrics used by the engine.
// types.go contains the main data structures for log payloads and processor interfaces.
package model

type LogPayload struct {
	Data         []byte
	FieldIndices []FieldIndex
}

type FieldIndex struct {
	KeyStart, KeyEnd int
	ValStart, ValEnd int
}

type Processor interface {
	Name() string
	Process(payload *LogPayload) (keep bool)
}
