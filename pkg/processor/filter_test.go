package processor

import (
	"testing"

	"github.com/loggling/loggling/pkg/model"
)

func TestFieldFilter_Process(t *testing.T) {

	tests := []struct {
		name             string
		filterKey        string
		filterVal        string
		inputData        string
		keyStart, keyEnd int
		valStart, valEnd int
		expected         bool
	}{
		{
			name:      "로그 레벨이 DEBUG인 경우 드롭(false)",
			filterKey: "level",
			filterVal: "DEBUG",
			inputData: `{"level":"DEBUG"}`,
			keyStart:  2, keyEnd: 7, // "level"의 위치
			valStart: 10, valEnd: 15, // "DEBUG"의 위치
			expected: false,
		},
		{
			name:      "로그 레벨이 INFO인 경우 유지(true)",
			filterKey: "level",
			filterVal: "DEBUG",
			inputData: `{"level":"INFO"}`,
			keyStart:  2, keyEnd: 7, // "level"의 위치
			valStart: 10, valEnd: 14, // "INFO"의 위치
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FieldFilter{
				TargetField: []byte(tt.filterKey),
				Value:       []byte(tt.filterVal),
			}

			payload := &model.LogPayload{
				Data: []byte(tt.inputData),
				FieldIndices: []model.FieldIndex{
					{
						KeyStart: tt.keyStart, KeyEnd: tt.keyEnd,
						ValStart: tt.valStart, ValEnd: tt.valEnd,
					},
				},
			}

			result := f.Process(payload)

			if result != tt.expected {
				t.Errorf("테스트 실패 [%s]: 예상값 %v, 실제값 %v", tt.name, tt.expected, result)
			}

		})
	}
}
