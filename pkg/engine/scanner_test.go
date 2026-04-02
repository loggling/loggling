package engine

import (
	"testing"

	"github.com/loggling/loggling/pkg/model"
)

func TestScanJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
		count   int
	}{
		{
			name:    "정상적인 JSON 파싱",
			json:    `{"name":"loggling","v":1}`,
			wantErr: false,
			count:   2,
		},
		{
			name:    "비정상적인 JSON 파싱",
			json:    `{"name":"loggling`,
			wantErr: true,
			count:   0,
		},
		{
			name:    "빈 JSON 객체",
			json:    `{}`,
			wantErr: false,
			count:   0,
		},
		{
			name:    "비정상적으로 빈 JSON 객체",
			json:    `{`,
			wantErr: true,
			count:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := &model.LogPayload{
				Data: []byte(tt.json),
			}

			err := ScanJSON(payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("ScanJSON() 에러 발생 여부 불일치: %v", err)
			}

			if len(payload.FieldIndices) != tt.count {
				t.Errorf("필드 개수 불일치: 예상 %d, 실제 %d", tt.count, len(payload.FieldIndices))

			}
		})
	}
}

func TestScanJSON_Details(t *testing.T) {

	jsonInput := `{"level":"INFO","msg":"hello world"}`
	payload := model.LogPayload{
		Data: []byte(jsonInput),
	}

	err := ScanJSON(&payload)

	if err != nil {
		t.Fatalf("기본 파싱 실패: %v", err)
	}

	firstField := payload.FieldIndices[0]

	key := string(payload.Data[firstField.KeyStart:firstField.KeyEnd])
	val := string(payload.Data[firstField.ValStart:firstField.ValEnd])

	if key != "level" || val != "INFO" {
		t.Errorf("첫 번째 필드 파싱 오류: 키(%s), 값(%s)", key, val)
	}

	secondField := payload.FieldIndices[1]

	key2 := string(payload.Data[secondField.KeyStart:secondField.KeyEnd])
	val2 := string(payload.Data[secondField.ValStart:secondField.ValEnd])

	if key2 != "msg" || val2 != "hello world" {
		t.Errorf("두 번째 필드 파싱 오류: 키(%s), 값(%s)", key, val)
	}
}
