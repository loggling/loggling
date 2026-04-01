package engine

import (
	"errors"

	"github.com/loggling/loggling/pkg/model"
)

func ScanJSON(payload *model.LogPayload) error {
	payload.FieldIndices = payload.FieldIndices[:0]
	data := payload.Data

	n := len(data)

	var current model.FieldIndex
	inQuotes := false
	waitingForValue := false

	for i := range n {
		b := data[i]

		switch b {
		case '"':
			inQuotes = !inQuotes
			if inQuotes {
				if waitingForValue {
					current.ValStart = i + 1
				} else {
					current.KeyStart = i + 1
				}
			} else {
				if !waitingForValue {
					current.KeyEnd = i
				} else {
					current.ValEnd = i
					payload.FieldIndices = append(payload.FieldIndices, current)
					waitingForValue = false
				}
			}
		case ':':
			if !inQuotes {
				waitingForValue = true

				j := i + 1
				for j < n && (data[j] == ' ' || data[j] == '\t') {
					j++
				}

				if j < n && data[j] != '"' {
					current.ValStart = j
				}
			}
		case ',', '}':
			if !inQuotes && waitingForValue {
				if current.ValStart > 0 && current.ValEnd <= current.ValStart {
					current.ValEnd = i
					payload.FieldIndices = append(payload.FieldIndices, current)
				}
				waitingForValue = false
			}
		}
	}

	if inQuotes {
		return errors.New("unclosed string quotes in JSON payload")
	}

	if len(payload.FieldIndices) == 0 && string(data) != "{}" {
		return errors.New("malformed JSON payload: no fields detected")
	}

	return nil
}
