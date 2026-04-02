// Package engine provides the core log processing mechanics.
// rotator.go provides thread-safe, rotatable file writers for output and DLQ logging.
package engine

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/loggling/loggling/pkg/model/logger"
)

type RotatableWriter struct {
	mu       sync.RWMutex
	filepath string
	file     *os.File
}

func NewRotatableWriter(filePath string) (*RotatableWriter, error) {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, err
	}

	rw := &RotatableWriter{
		filepath: filePath,
	}

	if err := rw.Rotate(); err != nil {
		return nil, err
	}

	return rw, nil
}

func (rw *RotatableWriter) Write(p []byte) (n int, err error) {
	rw.mu.RLock()
	defer rw.mu.RUnlock()
	return rw.file.Write(p)
}

func (rw *RotatableWriter) Rotate() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if rw.file != nil {
		rw.file.Close()
	}

	f, err := os.OpenFile(rw.filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		logger.Error("[Rotator] Failed to open log file:", err)
		return err
	}

	rw.file = f
	logger.Info("[Rotator] Log file successfully rotated:", rw.filepath)

	return nil
}

func (rw *RotatableWriter) Close() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if rw.file != nil {
		return rw.file.Close()
	}

	return nil
}
