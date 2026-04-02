// Package processor provides log data transformation logic.
// deduplication.go filters out identical consecutive log entries to reduce noise.
package processor

import (
	"bytes"

	"github.com/loggling/loggling/pkg/model"
)

type DeduplicationProcessor struct {
	lastLog []byte
}

func (d *DeduplicationProcessor) Name() string {
	return "JSON_DEDUPLICATION"
}

func (d *DeduplicationProcessor) Process(payload *model.LogPayload) bool {
	if bytes.Equal(d.lastLog, payload.Data) {
		return false
	}

	d.lastLog = append(d.lastLog[:0], payload.Data...)
	return true
}
