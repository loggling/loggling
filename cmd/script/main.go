// Package main provides a utility script for generating large volumes of test log data.
package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/loggling/loggling/pkg/model/logger"
)

func main() {
	const fileName = "input.log"
	const targetLines = 10_000_000

	logger.Default("INFO")

	file, err := os.Create(fileName)
	if err != nil {
		logger.Error("Failed to create log file:", err)
		os.Exit(1)
	}
	defer file.Close()

	writer := bufio.NewWriterSize(file, 1024*1024)
	defer writer.Flush()

	levels := []string{"INFO", "DEBUG", "ERROR", "WARN"}
	users := []string{"james", "anna", "admin", "guest", "bot_01"}

	logger.Info("Starting generation of", fileName, "(Target:", targetLines, "lines)...")
	start := time.Now()

	for i := range targetLines {
		logLine := fmt.Sprintf(
			`{"id":%d,"ts":"%s","level":"%s","user":"%s","password":"pwd_%d","msg":"user logged in from 127.0.0.1"}`+"\n",
			i,
			time.Now().Format(time.RFC3339),
			levels[rand.Intn(len(levels))],
			users[rand.Intn(len(users))],
			rand.Intn(1000000),
		)

		_, err := writer.WriteString(logLine)
		if err != nil {
			logger.Error("Failed to write to log file:", err)
			os.Exit(1)
		}

		if i%1_000_000 == 0 && i > 0 {
			logger.Info("... Completed", i/1_000_000, "million lines")
		}
	}

	logger.Info("Generation complete! Elapsed time:", time.Since(start))
}
