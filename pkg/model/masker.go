// Package model defines the core data structures and metrics used by the engine.
// masker.go defines the configuration for sensitive field masking.
package model

import (
	"bytes"
)

type MaskStrategy interface {
	Mask(data []byte)
}

type FixedMasker struct {
}

type PartialMasker struct {
	Delimiter  byte
	KeepPrefix int
	KeepSuffix int
}

type SegmentMasker struct {
	Delimiter byte
}

func (f *FixedMasker) Mask(data []byte) {
	for i := range data {
		data[i] = '*'
	}
}

func (p *PartialMasker) Mask(data []byte) {
	if p.Delimiter != 0 {
		idx := bytes.IndexByte(data, p.Delimiter)

		if idx != -1 {
			p.maskRange(data[:idx], p.KeepPrefix, p.KeepSuffix)
			return
		}
	}

	p.maskRange(data, p.KeepPrefix, p.KeepSuffix)
}

func (p *PartialMasker) maskRange(data []byte, keepPrefix, keepSuffix int) {
	start := keepPrefix
	end := len(data) - keepSuffix

	if start >= end || start < 0 {
		return
	}

	for i := start; i < end; i++ {
		data[i] = '*'
	}
}

func (s *SegmentMasker) Mask(data []byte) {
	firstDash := bytes.IndexByte(data, s.Delimiter)
	lastDash := bytes.LastIndexByte(data, s.Delimiter)
	if firstDash == -1 || firstDash == lastDash {
		return
	}
	for i := firstDash + 1; i < lastDash; i++ {
		if data[i] != s.Delimiter {
			data[i] = '*'
		}
	}
}
