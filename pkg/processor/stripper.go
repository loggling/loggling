package processor

import (
	"bytes"

	"github.com/loggling/loggling/pkg/model"
)

type FieldStripper struct {
	TargetFields [][]byte
}

func (s *FieldStripper) Name() string {
	return "FIELD_STRIPPER"
}

func (s *FieldStripper) Process(payload *model.LogPayload) bool {
	data := payload.Data

	for i := len(payload.FieldIndices) - 1; i >= 0; i-- {
		idx := payload.FieldIndices[i]
		key := data[idx.KeyStart:idx.KeyEnd]

		shouldDelete := false
		for _, target := range s.TargetFields {
			if bytes.Equal(key, target) {
				shouldDelete = true
				break
			}
		}

		if shouldDelete {

			if idx.KeyStart <= 0 || idx.ValEnd <= idx.KeyStart {
				continue
			}

			start := idx.KeyStart - 1
			end := idx.ValEnd + 1

			if start > 0 && data[start-1] == ',' {
				start--
			} else if end < len(data) && data[end] == ',' {
				end++
			}

			copy(data[start:], data[end:])
			data = data[:len(data)-(end-start)]
		}

	}

	payload.Data = data
	return true
}
