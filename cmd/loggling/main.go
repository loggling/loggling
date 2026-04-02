// Package main is the entry point for the Loggling daemon.
// It initializes configurations, sets up the log processing pipeline,
// and starts either the background file processors or the real-time HTTP gateway.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/loggling/loggling/pkg/config"
	"github.com/loggling/loggling/pkg/engine"
	"github.com/loggling/loggling/pkg/model/logger"
	"github.com/loggling/loggling/pkg/server"
)

func main() {
	godotenv.Load("../.env.local")
	log_level := os.Getenv("LOG_LEVEL")
	logger.Default(log_level)

	configPath := flag.String("config", "./configs/config.yaml", "path to config.yaml file")
	showVersion := flag.Bool("version", false, "show version and exit")
	fmt.Printf("\033[1;34m")
	fmt.Println(`                                          `)
	fmt.Println(`  _                      _ _               `)
	fmt.Println(` | |    ___   __ _  __ _| (_)_ __   __ _ `)
	fmt.Println(` | |   / _ \ / _` + "`" + ` |/ _` + "`" + ` | | | '_ \ / _` + "`" + ` |`)
	fmt.Println(` | |__| (_) | (_| | (_| | | | | | | (_| |`)
	fmt.Println(` |_____\___/ \__, |\__, |_|_|_| |_|\__, |`)
	fmt.Println(`             |___/ |___/            |___/  v0.5.3`)
	fmt.Println(`                                          `)
	fmt.Printf("\033[0m")
	flag.Usage = func() {

		fmt.Println("\nHigh-performance, parallel log gateway for cloud & local pipelines.\n")

		fmt.Println("\033[1mUSAGE:\033[0m")
		fmt.Println("    loggling [OPTIONS]")

		fmt.Println("\n\033[1mOPTIONS:\033[0m")
		fmt.Printf("    %-16s %s\n", "-config", "Path to config.yaml file (default: \"./configs/config.yaml\")")
		fmt.Printf("    %-16s %s\n", "-version", "Show version information and exit")
		fmt.Printf("    %-16s %s\n", "-h, --help", "Show this professional help message")

		fmt.Println("\n\033[1mEXAMPLES:\033[0m")
		fmt.Println("    $ \033[32mloggling -config configs/config.yaml\033[0m    Run with custom configuration")
		fmt.Println("    $ \033[32mloggling -version\033[0m                 Check current engine version")

		fmt.Println("\nFor more information, visit https://github.com/loggling/loggling")
	}

	flag.Parse()

	if *showVersion {
		logger.Info("[Loggling] v0.5.3")
		return
	}

	cfg, err := config.LoadConfig(*configPath)

	if err != nil {
		logger.Error("[Loggling] Failed to load config from", *configPath, ":", err)
		os.Exit(1)
	}

	pipe := engine.NewPipelineFromConfig(cfg)
	var targetFiles []string

	absOutput, _ := filepath.Abs(filepath.Clean(cfg.Default.Output))
	absDLQ, _ := filepath.Abs(filepath.Clean(cfg.Default.DLQ))

	for _, pattern := range cfg.Default.Inputs {
		matches, err := filepath.Glob(pattern)
		if err == nil {

			for _, match := range matches {
				absMatch, _ := filepath.Abs(filepath.Clean(match))

				if absMatch == absOutput || absMatch == absDLQ {
					logger.Debug("Excluding output/dlq file:", absMatch)
					continue
				}
				targetFiles = append(targetFiles, match)
			}
		}
	}

	if !cfg.Server.Enabled && len(targetFiles) == 0 {
		logger.Error("empty files (pattern:", cfg.Default.Inputs, ")")
		os.Exit(1)
	}

	outFile, err := engine.NewRotatableWriter(cfg.Default.Output)

	if err != nil {
		logger.Error("output file error:", err)
		os.Exit(1)
	}

	defer outFile.Close()

	var dlqFile *engine.RotatableWriter

	if cfg.Default.DLQ != "" {
		var err error
		dlqFile, err = engine.NewRotatableWriter(cfg.Default.DLQ)
		if err != nil {
			logger.Error("dlq file error:", err)
			os.Exit(1)
		}

		defer dlqFile.Close()
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)

	go func() {
		for sig := range sigChan {
			logger.Info("[Loggling] Caught OS signal (", sig, "): rotating log file handle...")
			outFile.Rotate()

			if dlqFile != nil {
				dlqFile.Rotate()
			}
		}
	}()

	logger.Info("[Loggling] Initialization sequence complete.")
	logger.Info("[Loggling] Configuration:", *configPath)
	logger.Info("[Loggling] Output Target:", cfg.Default.Output)
	if cfg.Default.DLQ != "" {
		logger.Info("[Loggling] Error DLQ Path:", cfg.Default.DLQ)
	}

	if cfg.Server.Enabled {
		logger.Info("[Loggling] Mode: Hybrid (Gateway :", cfg.Server.Port, "+ Local File)")
	} else {
		logger.Info("[Loggling] Mode: Standalone File Processor")
	}

	runner := engine.NewStreamRunner(pipe, dlqFile)

	go config.WatchConfig(*configPath, 2*time.Second, func(newCfg *config.Config) {
		newPipeline := engine.NewPipelineFromConfig(newCfg)
		runner.SwapPipeline(newPipeline)
		logger.Info("[Loggling] Hot-reload complete! New YAML ruleset applied without downtime.")
	})

	processLocalFiles := func() {
		var inputs []io.Reader
		for _, file := range targetFiles {
			inFile, err := os.Open(file)
			if err != nil {
				logger.Warn("[", file, "] failed to open:", err)
				continue
			}
			inputs = append(inputs, inFile)
		}

		if len(inputs) == 0 {
			return
		}

		defer func() {
			for _, r := range inputs {
				if f, ok := r.(*os.File); ok {
					f.Close()
				}
			}
		}()

		logger.Info("[Loggling] Starting to process", len(inputs), "local files...")
		if err := runner.RunParallel(inputs, targetFiles, outFile); err != nil {
			logger.Error("[Loggling] Runtime error during parallel processing:", err)
		}
		logger.Info("[Loggling] Local file processing complete.")
	}

	if cfg.Server.Enabled {
		go processLocalFiles()

		gateway := &server.Gateway{
			Runner: runner,
			Output: outFile,
			Port:   cfg.Server.Port,
		}

		if err := gateway.Start(); err != nil {
			logger.Error("[Gateway] Failed to start server:", err)
			os.Exit(1)
		}
	} else {
		processLocalFiles()
	}
}
