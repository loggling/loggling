package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

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

	var dlqFile *engine.RotatableWriter

	if cfg.Default.DLQ != "" {
		var err error
		dlqFile, err = engine.NewRotatableWriter(cfg.Default.DLQ)
		if err != nil {
			log.Fatalf("dlq file error: %v", err)
		}

		defer dlqFile.Close()
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)

	go func() {
		for sig := range sigChan {
			log.Printf("[Loggling] Caught OS signal (%v): rotating log file handle...", sig)
			outFile.Rotate()

			if dlqFile != nil {
				dlqFile.Rotate()
			}
		}
	}()

	runner := engine.NewStreamRunner(pipe, dlqFile)

	go config.WatchConfig("./configs/config.yaml", 2*time.Second, func(newCfg *config.Config) {
		newPipeline := engine.NewPipelineFromConfig(newCfg)
		// Atomically swap the pipeline without any locks
		runner.SwapPipeline(newPipeline)
		log.Println("[Loggling] Hot-reload complete! New YAML ruleset applied without downtime.")
	})

	// 로컬 파일 처리 로직을 익명 함수로 묶습니다.
	processLocalFiles := func() {
		var inputs []io.Reader
		for _, file := range targetFiles {
			inFile, err := os.Open(file)
			if err != nil {
				log.Printf("[%s] failed to open: %v", file, err)
				continue
			}
			inputs = append(inputs, inFile)
		}

		if len(inputs) == 0 {
			return
		}

		// 파일 처리 종료 시 열어둔 파일들을 닫습니다.
		defer func() {
			for _, r := range inputs {
				if f, ok := r.(*os.File); ok {
					f.Close()
				}
			}
		}()

		log.Printf("Starting to process %d local files...", len(inputs))
		if err := runner.RunParallel(inputs, outFile); err != nil {
			log.Printf("runtime error: %v", err)
		}
		fmt.Println("Local file processing complete.")
	}

	if cfg.Server.Enabled {
		// 💥 핵심: 게이트웨이가 켜져 있다면, 로컬 파일 처리는 백그라운드(Goroutine)로 돌립니다!
		go processLocalFiles()

		// 게이트웨이(HTTP 서버)는 메인 스레드에서 무한 대기하며 실행됩니다.
		gateway := &server.Gateway{
			Runner: runner,
			Output: outFile,
			Port:   cfg.Server.Port,
		}

		if err := gateway.Start(); err != nil {
			log.Fatalf("Failed to start gateway server: %v", err)
		}
	} else {
		// 서버 모드가 아닐 경우 기존처럼 메인 스레드에서 파일을 처리하고 끝냅니다.
		processLocalFiles()
	}
}
