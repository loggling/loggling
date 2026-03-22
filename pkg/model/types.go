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
