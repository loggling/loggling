package processor

import (
	"bytes"

	"github.com/loggling/loggling/pkg/model"
)

type FieldFilter struct {
	TargetField []byte
	Value       []byte
}

func (f *FieldFilter) Name() string {
	return "FIELD_FILTER"
}

func (f *FieldFilter) Process(payload *model.LogPayload) bool {
	for _, idx := range payload.FieldIndices {
		key := payload.Data[idx.KeyStart:idx.KeyEnd]
		if bytes.Equal(key, f.TargetField) {
			val := payload.Data[idx.ValStart:idx.ValEnd]

			if bytes.Equal(val, f.Value) {
				return false
			}

			return true
		}
	}
	return true
}
