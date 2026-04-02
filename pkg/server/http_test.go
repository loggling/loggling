package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/loggling/loggling/pkg/engine"
)

func TestGateway_HandleLogs(t *testing.T) {

	pipe := engine.NewPipeline()
	runner := engine.NewStreamRunner(pipe, nil)

	var out bytes.Buffer

	g := &Gateway{
		Runner: runner,
		Output: &out,
		Port:   8080,
	}

	logData := `{"msg":"test network log"}`

	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewBufferString(logData))

	w := httptest.NewRecorder()

	g.handleLogs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("처리 실패: 상태 코드 %d", w.Code)
	}

	reqGet := httptest.NewRequest(http.MethodGet, "/logs", nil)
	wGet := httptest.NewRecorder()

	g.handleLogs(wGet, reqGet)
	if wGet.Code != http.StatusMethodNotAllowed {
		t.Errorf("잘못된 메소드 차단 실패: 상태 코드 %d", wGet.Code)
	}
}
