package engine

import (
	"strconv"

	"github.com/loggling/loggling/pkg/config"
	"github.com/loggling/loggling/pkg/model"
	"github.com/loggling/loggling/pkg/processor"
)

func NewPipelineFromConfig(cfg *config.Config) *Pipeline {
	paramConfigs := cfg.Pipeline
	processors := make([]model.Processor, 0)

	if items, ok := paramConfigs["filter"]; ok {
		for _, item := range items {
			processors = append(processors, &processor.FieldFilter{
				TargetField: []byte(item["field"]),
				Value:       []byte(item["value"]),
			})
		}
	}

	if items, ok := paramConfigs["stripper"]; ok {
		for _, item := range items {
			targets := make([][]byte, 0, len(item))
			for k, v := range item {
				if k == "field" {
					targets = append(targets, []byte(v))
				}
			}

			processors = append(processors, &processor.FieldStripper{
				TargetFields: targets,
			})
		}
	}

	if items, ok := paramConfigs["masker"]; ok {
		for _, item := range items {
			strategy := createMaskStrategy(item["preset"], item)
			processors = append(processors, &processor.JsonMasker{
				TargetField: []byte(item["field"]),
				Strategy:    strategy,
			})
		}
	}
	processors = append(processors, &processor.DeduplicationProcessor{})

	return NewPipeline(processors...)
}

func createMaskStrategy(preset string, params map[string]string) model.MaskStrategy {
	switch preset {
	case "password":
		return &model.PartialMasker{}
	case "email":
		return &model.PartialMasker{Delimiter: '@', KeepPrefix: 1}
	case "card":
		return &model.SegmentMasker{Delimiter: '-'}
	}

	keepPrefix, _ := strconv.Atoi(params["keep_prefix"])
	keepSuffix, _ := strconv.Atoi(params["keep_suffix"])
	var delimiter byte
	if d, ok := params["delimiter"]; ok && len(d) > 0 {
		delimiter = d[0]
	}

	return &model.PartialMasker{
		Delimiter:  delimiter,
		KeepPrefix: keepPrefix,
		KeepSuffix: keepSuffix,
	}
}
