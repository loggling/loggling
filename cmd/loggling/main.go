package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/loggling/loggling/pkg/config"
	"github.com/loggling/loggling/pkg/engine"
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

	if len(targetFiles) == 0 {
		log.Fatalf("empty files (패턴: %v)", cfg.Default.Inputs)
	}

	outFile, err := os.Create(cfg.Default.Output)
	if err != nil {
		log.Fatalf("output file error: %v", err)
	}
	defer outFile.Close()

	runner := &engine.StreamRunner{Pipeline: pipe}
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
