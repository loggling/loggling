package config

import (
	"log"
	"os"
	"time"
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

			log.Printf("[Hot-Reload] Detected changes in '%s'. Compiling new ruleset...", filePath)

			newCfg, err := LoadConfig(filePath)
			if err != nil {
				log.Printf("[Hot-Reload] WARNING: Syntax error in YAML. Keeping the existing configuration intact. (%v)", err)
				continue
			}

			onUpdate(newCfg)
		}
	}
}
