// Package server handles the network ingestion layer.
// http.go implements the POST /logs gateway for real-time log ingestion.
package server

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/loggling/loggling/pkg/engine"
)

type Gateway struct {
	Runner *engine.StreamRunner
	Output io.Writer
	Port   int
}

func (g *Gateway) Start() error {
	http.HandleFunc("/logs", g.handleLogs)

	addr := fmt.Sprintf(":%d", g.Port)
	log.Printf("🚀 Loggling Gateway Server Started on %s", addr)
	return http.ListenAndServe(addr, nil)
}

func (g *Gateway) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed (only POST)", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	if err := g.Runner.Run(r.Body, g.Output); err != nil {
		log.Printf("Request Parsing Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}
