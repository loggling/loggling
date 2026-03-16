package engine

import (
	"sync"

	"github.com/loggling/loggling/pkg/model"
)

type Pipeline struct {
	processors []model.Processor
	pool       sync.Pool
}

func NewPipeline(processors ...model.Processor) *Pipeline {
	return &Pipeline{
		processors: processors,
		pool: sync.Pool{
			New: func() any {
				return &model.LogPayload{
					Data:         make([]byte, 0, 4096),
					FieldIndices: make([]model.FieldIndex, 0, 16),
				}
			},
		},
	}
}

func (p *Pipeline) Execute(input []byte) *model.LogPayload {
	payload := p.pool.Get().(*model.LogPayload)
	payload.Data = payload.Data[:0]
	payload.Data = append(payload.Data, input...)

	ScanJSON(payload)

	for _, proc := range p.processors {
		if keep := proc.Process(payload); !keep {
			p.Release(payload)
			model.GlobalMetrics.AddDroppedLine()
			return nil
		}

		if proc.Name() == "FIELD_STRIPPER" {
			ScanJSON(payload)
		}
	}

	model.GlobalMetrics.AddProcessedLine(len(payload.Data))
	return payload
}

func (p *Pipeline) Release(payload *model.LogPayload) {
	p.pool.Put(payload)
}
