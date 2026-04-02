package processor

import (
	"testing"

	"github.com/loggling/loggling/pkg/model"
)

func TestStripper_Process(t *testing.T) {
	tests := []struct {
		name             string
		stripField       string
		inputData        string
		keyStart, keyEnd int
		valStart, valEnd int
		expectedData     string
	}{
		{
			name:         "첫 번째 필드 삭제 (콤마 처리 확인)",
			stripField:   "user",
			inputData:    `{"user":"admin","msg":"hi"}`,
			keyStart:     2,
			keyEnd:       6,
			valStart:     9,
			valEnd:       14,
			expectedData: `{"msg":"hi"}`,
		},
		{
			name:         "마지막 필드 삭제",
			stripField:   "msg",
			inputData:    `{"user":"admin","msg":"hi"}`,
			keyStart:     17,
			keyEnd:       20,
			valStart:     23,
			valEnd:       25,
			expectedData: `{"user":"admin"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &FieldStripper{
				TargetFields: [][]byte{[]byte(tt.stripField)},
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

			s.Process(payload)

			if string(payload.Data) != tt.expectedData {
				t.Errorf("삭제 실패: 결과 [%s], 예상 [%s]", string(payload.Data), tt.expectedData)
			}
		})
	}
}
