package config

import (
	"log"
	"os"
	"time"
)

// WatchConfig monitors the configuration file for modifications
// and triggers the onUpdate callback with a newly parsed Config object.
// It uses a zero-dependency polling mechanism via os.Stat.
func WatchConfig(filePath string, interval time.Duration, onUpdate func(*Config)) {
	var lastModTime time.Time

	// Record the initial file modification time
	state, err := os.Stat(filePath)
	if err == nil {
		lastModTime = state.ModTime()
	}

	ticker := time.NewTicker(interval)

	// Start the infinite polling loop in the background
	for range ticker.C {
		state, err := os.Stat(filePath)
		if err != nil {
			// Skip if the file is temporarily unavailable or deleted
			continue
		}

		currentModTime := state.ModTime()
		if currentModTime.After(lastModTime) {
			lastModTime = currentModTime

			log.Printf("[Hot-Reload] Detected changes in '%s'. Compiling new ruleset...", filePath)

			// Try to parse the updated YAML
			newCfg, err := LoadConfig(filePath)
			if err != nil {
				// Prevent crashes if the user provided invalid YAML syntax
				log.Printf("[Hot-Reload] WARNING: Syntax error in YAML. Keeping the existing configuration intact. (%v)", err)
				continue
			}

			// Successfully parsed; trigger the hot swap securely
			onUpdate(newCfg)
		}
	}
}
