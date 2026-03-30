package main

import (
	"fmt"
	"log"
	"os"

	"github.com/loggling/loggling/pkg/config"
	"github.com/loggling/loggling/pkg/engine"
)

func main() {
	cfg, err := config.LoadConfig("./configs/config.yaml")
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	pipe := engine.NewPipelineFromConfig(cfg)

	inFile, err := os.Open(cfg.Default.InputPath)
	if err != nil {
		log.Fatalf("input file error: %v", err)
	}
	defer inFile.Close()

	outFile, err := os.Create(cfg.Default.OutputPath)
	if err != nil {
		log.Fatalf("output file error: %v", err)
	}
	defer outFile.Close()

	runner := &engine.StreamRunner{Pipeline: pipe}
	if err := runner.Run(inFile, outFile); err != nil {
		log.Fatalf("runtime error: %v", err)
	}

	fmt.Println("Processing complete.")
}
