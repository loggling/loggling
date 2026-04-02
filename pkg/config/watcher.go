// Package config handles the initialization and hot-reloading of project settings.
// watcher.go utilizes a zero-dependency os.Stat polling mechanism to detect
// configuration file changes and trigger highly available, lock-free pipeline swaps.
package config

import (
	"os"
	"time"

	"github.com/loggling/loggling/pkg/model/logger"
)

func WatchConfig(filePath string, interval time.Duration, onUpdate func(*Config)) {
	var lastModTime time.Time

	state, err := os.Stat(filePath)
	if err == nil {
		lastModTime = state.ModTime()
	}

	ticker := time.NewTicker(interval)

	for range ticker.C {
		state, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		currentModTime := state.ModTime()
		if currentModTime.After(lastModTime) {
			lastModTime = currentModTime

			logger.Info("[Hot-Reload] Detected changes in '", filePath, "'. Compiling new ruleset...")

			newCfg, err := LoadConfig(filePath)
			if err != nil {
				logger.Warn("[Hot-Reload] WARNING: Syntax error in YAML. Keeping the existing configuration intact. (", err, ")")
				continue
			}

			onUpdate(newCfg)
		}
	}
}
