package processor

import (
	"bytes"
	"testing"

	"github.com/loggling/loggling/pkg/model"
)

func TestJsonMasker_Process(t *testing.T) {
	tests := []struct {
		name             string
		maskField        string
		inputData        string
		keyStart, keyEnd int
		valStart, valEnd int
		expectedData     string
	}{
		{
			name:      "비밀번호 필드 전체 마스킹",
			maskField: "password",
			inputData: `{"password":"secret123"}`,
			keyStart:  2, keyEnd: 10,
			valStart: 13, valEnd: 22,
			expectedData: `{"password":"*********"}`, // FixedMasker 사용 시
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := &model.FixedMasker{}
			m := &JsonMasker{
				TargetField: []byte(tt.maskField),
				Strategy:    strategy,
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

			m.Process(payload)

			if !bytes.Equal(payload.Data, []byte(tt.expectedData)) {
				t.Errorf("마스킹 실패: 결과 %s, 예상 %s", string(payload.Data), tt.expectedData)
			}
		})
	}
}
