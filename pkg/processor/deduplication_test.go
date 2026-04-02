package processor

import (
	"testing"

	"github.com/loggling/loggling/pkg/model"
)

func TestDeduplicationProcessor_Process(t *testing.T) {
	d := &DeduplicationProcessor{}

	log1 := `{"id":1,"msg":"error occurred"}`
	log2 := `{"id":2,"msg":"another message"}`

	p1 := &model.LogPayload{
		Data: []byte(log1),
	}

	if !d.Process(p1) {
		t.Error("새로운 로그 유입 시 true를 반환해야 합니다.")
	}

	p2 := &model.LogPayload{
		Data: []byte(log1),
	}

	if d.Process(p2) {
		t.Error("중복 로그 유입 시 false를 반환해야 합니다.")

	}

	// 4. [세 번째 시도] 다른 내용의 로그가 들어왔을 때 -> 결과는 다시 true(유지)여야 함
	p3 := &model.LogPayload{Data: []byte(log2)}
	if !d.Process(p3) {
		t.Error("다른 내용의 로그 유입 시 true를 반환해야 합니다.")
	}

}
