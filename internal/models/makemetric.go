package models

import (
	"github.com/lavaorg/telex"
)

// Makemetric applies new metric plugin and agent measurement and tag
// settings.
func makemetric(
	metric telex.Metric,
	nameOverride string,
	namePrefix string,
	nameSuffix string,
	tags map[string]string,
	globalTags map[string]string,
) telex.Metric {
	if len(nameOverride) != 0 {
		metric.SetName(nameOverride)
	}

	if len(namePrefix) != 0 {
		metric.AddPrefix(namePrefix)
	}
	if len(nameSuffix) != 0 {
		metric.AddSuffix(nameSuffix)
	}

	// Apply plugin-wide tags
	for k, v := range tags {
		if _, ok := metric.GetTag(k); !ok {
			metric.AddTag(k, v)
		}
	}
	// Apply global tags
	for k, v := range globalTags {
		if _, ok := metric.GetTag(k); !ok {
			metric.AddTag(k, v)
		}
	}

	return metric
}
