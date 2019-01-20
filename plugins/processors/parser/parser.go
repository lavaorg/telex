package parser

import (
	"log"

	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/plugins/parsers"
	"github.com/lavaorg/telex/plugins/processors"
)

type Parser struct {
	parsers.Config
	DropOriginal bool     `toml:"drop_original"`
	Merge        string   `toml:"merge"`
	ParseFields  []string `toml:"parse_fields"`
	Parser       parsers.Parser
}

func (p *Parser) Apply(metrics ...telex.Metric) []telex.Metric {
	if p.Parser == nil {
		var err error
		p.Parser, err = parsers.NewParser(&p.Config)
		if err != nil {
			log.Printf("E! [processors.parser] could not create parser: %v", err)
			return metrics
		}
	}

	results := []telex.Metric{}

	for _, metric := range metrics {
		newMetrics := []telex.Metric{}
		if !p.DropOriginal {
			newMetrics = append(newMetrics, metric)
		}

		for _, key := range p.ParseFields {
			for _, field := range metric.FieldList() {
				if field.Key == key {
					switch value := field.Value.(type) {
					case string:
						fromFieldMetric, err := p.parseField(value)
						if err != nil {
							log.Printf("E! [processors.parser] could not parse field %s: %v", key, err)
						}

						for _, m := range fromFieldMetric {
							if m.Name() == "" {
								m.SetName(metric.Name())
							}
						}

						// multiple parsed fields shouldn't create multiple
						// metrics so we'll merge tags/fields down into one
						// prior to returning.
						newMetrics = append(newMetrics, fromFieldMetric...)
					default:
						log.Printf("E! [processors.parser] field '%s' not a string, skipping", key)
					}
				}
			}
		}

		if len(newMetrics) == 0 {
			continue
		}

		if p.Merge == "override" {
			results = append(results, merge(newMetrics[0], newMetrics[1:]))
		} else {
			results = append(results, newMetrics...)
		}
	}
	return results
}

func merge(base telex.Metric, metrics []telex.Metric) telex.Metric {
	for _, metric := range metrics {
		for _, field := range metric.FieldList() {
			base.AddField(field.Key, field.Value)
		}
		for _, tag := range metric.TagList() {
			base.AddTag(tag.Key, tag.Value)
		}
		base.SetName(metric.Name())
	}
	return base
}

func (p *Parser) parseField(value string) ([]telex.Metric, error) {
	return p.Parser.Parse([]byte(value))
}

func init() {
	processors.Add("parser", func() telex.Processor {
		return &Parser{DropOriginal: false}
	})
}
