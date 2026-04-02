// Package server handles the network ingestion layer.
// http.go implements the POST /logs gateway for real-time log ingestion.
package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/loggling/loggling/pkg/engine"
	"github.com/loggling/loggling/pkg/model/logger"
)

type Gateway struct {
	Runner *engine.StreamRunner
	Output io.Writer
	Port   int
}

func (g *Gateway) Start() error {
	http.HandleFunc("/logs", g.handleLogs)
	http.HandleFunc("/health", g.handleHealth)

	addr := fmt.Sprintf(":%d", g.Port)
	logger.Info("[Gateway] Started Loggling HTTP Gateway on", addr)
	return http.ListenAndServe(addr, nil)
}
func (g *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}
func (g *Gateway) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed (only POST)", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	if err := g.Runner.Run(r.Body, g.Output); err != nil {
		logger.Error("[Gateway] Request Parsing Error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}
