package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/loggling/loggling/pkg/config"
	"github.com/loggling/loggling/pkg/engine"
	"github.com/loggling/loggling/pkg/server"
)

func main() {
	cfg, err := config.LoadConfig("./configs/config.yaml")
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	pipe := engine.NewPipelineFromConfig(cfg)
	var targetFiles []string

	for _, pattern := range cfg.Default.Inputs {
		matches, err := filepath.Glob(pattern)
		if err == nil {
			targetFiles = append(targetFiles, matches...)
		}
	}

	if !cfg.Server.Enabled && len(targetFiles) == 0 {
		log.Fatalf("empty files (pattern: %v)", cfg.Default.Inputs)
	}

	outFile, err := engine.NewRotatableWriter(cfg.Default.Output)

	if err != nil {
		log.Fatalf("output file error: %v", err)
	}

	defer outFile.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)

	go func() {
		for sig := range sigChan {
			log.Printf("[Loggling] Caught OS signal (%v): rotating log file handle...", sig)
			outFile.Rotate()
		}
	}()

	runner := &engine.StreamRunner{Pipeline: pipe}
	if cfg.Server.Enabled {
		gateway := &server.Gateway{
			Runner: runner,
			Output: outFile,
			Port:   cfg.Server.Port,
		}

		if err := gateway.Start(); err != nil {
			log.Fatalf("Failed to start gateway server: %v", err)
		}

		return
	}

	var inputs []io.Reader

	for _, file := range targetFiles {
		inFile, err := os.Open(file)

		if err != nil {
			log.Printf("[%s] failed to open: %v", file, err)
			continue
		}

		defer inFile.Close()
		inputs = append(inputs, inFile)
	}

	if err := runner.RunParallel(inputs, outFile); err != nil {
		log.Printf("runtime error: %v", err)
	}

	fmt.Println("Processing complete.")
}
