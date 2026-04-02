// Package processor provides log data transformation logic.
// masker.go sensitive fields within the log data (e.g. passwords).
package processor

import (
	"bytes"

	"github.com/loggling/loggling/pkg/model"
)

type JsonMasker struct {
	TargetField []byte
	Strategy    model.MaskStrategy
}

func (m *JsonMasker) Name() string {
	return "JSON_MASKER"
}

func (m *JsonMasker) Process(payload *model.LogPayload) bool {
	for _, idx := range payload.FieldIndices {
		key := payload.Data[idx.KeyStart:idx.KeyEnd]

		if bytes.Equal(key, m.TargetField) {
			valueRange := payload.Data[idx.ValStart:idx.ValEnd]
			m.Strategy.Mask(valueRange)

			return true
		}
	}

	return true
}
