package engine

import (
	"bytes"
	"testing"

	"github.com/loggling/loggling/pkg/model"
	"github.com/loggling/loggling/pkg/processor"
)

func TestPipeline_Execute(t *testing.T) {
	filter := &processor.FieldFilter{
		TargetField: []byte("level"),
		Value:       []byte("DEBUG"),
	}

	masker := &processor.JsonMasker{
		TargetField: []byte("val"),
		Strategy:    &model.FixedMasker{},
	}

	p := NewPipeline(filter, masker)

	tests := []struct {
		name     string
		input    string
		isDrop   bool
		expected string
	}{
		{
			name:     "일반 로그 - 마스킹만 적용됨",
			input:    `{"level":"INFO","val":"secret"}`,
			isDrop:   false,
			expected: `{"level":"INFO","val":"******"}`,
		},
		{
			name:     "DEBUG 로그 - 필터에 의해 드롭됨",
			input:    `{"level":"DEBUG","val":"secret"}`,
			isDrop:   true,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := p.Execute([]byte(tt.input))

			if err != nil {
				t.Fatalf("실행 중 에러 발생: %v", err)
			}

			if tt.isDrop {
				if res != nil {
					t.Error("로그가 드롭되어야 하는데 결과가 반환되었습니다.")
				}
				return
			}

			if !bytes.Equal(res.Data, []byte(tt.expected)) {
				t.Errorf("최종 데이터 불일치: 결과 %s, 예상 %s", string(res.Data), tt.expected)

			}
		})
	}
}
