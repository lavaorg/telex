package enum

import (
	"strconv"

	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/plugins/processors"
)

type EnumMapper struct {
	Mappings []Mapping `toml:"mapping"`
}

type Mapping struct {
	Field         string
	Dest          string
	Default       interface{}
	ValueMappings map[string]interface{}
}

func (mapper *EnumMapper) Apply(in ...telex.Metric) []telex.Metric {
	for i := 0; i < len(in); i++ {
		in[i] = mapper.applyMappings(in[i])
	}
	return in
}

func (mapper *EnumMapper) applyMappings(metric telex.Metric) telex.Metric {
	for _, mapping := range mapper.Mappings {
		if originalValue, isPresent := metric.GetField(mapping.Field); isPresent == true {
			if adjustedValue, isString := adjustBoolValue(originalValue).(string); isString == true {
				if mappedValue, isMappedValuePresent := mapping.mapValue(adjustedValue); isMappedValuePresent == true {
					writeField(metric, mapping.getDestination(), mappedValue)
				}
			}
		}
	}
	return metric
}

func adjustBoolValue(in interface{}) interface{} {
	if mappedBool, isBool := in.(bool); isBool == true {
		return strconv.FormatBool(mappedBool)
	}
	return in
}

func (mapping *Mapping) mapValue(original string) (interface{}, bool) {
	if mapped, found := mapping.ValueMappings[original]; found == true {
		return mapped, true
	}
	if mapping.Default != nil {
		return mapping.Default, true
	}
	return original, false
}

func (mapping *Mapping) getDestination() string {
	if mapping.Dest != "" {
		return mapping.Dest
	}
	return mapping.Field
}

func writeField(metric telex.Metric, name string, value interface{}) {
	metric.RemoveField(name)
	metric.AddField(name, value)
}

func init() {
	processors.Add("enum", func() telex.Processor {
		return &EnumMapper{}
	})
}
