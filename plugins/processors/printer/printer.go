package printer

import (
	"fmt"

	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/plugins/processors"
	"github.com/lavaorg/telex/plugins/serializers"
	"github.com/lavaorg/telex/plugins/serializers/influx"
)

type Printer struct {
	serializer serializers.Serializer
}

var sampleConfig = `
`

func (p *Printer) SampleConfig() string {
	return sampleConfig
}

func (p *Printer) Description() string {
	return "Print all metrics that pass through this filter."
}

func (p *Printer) Apply(in ...telex.Metric) []telex.Metric {
	for _, metric := range in {
		octets, err := p.serializer.Serialize(metric)
		if err != nil {
			continue
		}
		fmt.Printf("%s", octets)
	}
	return in
}

func init() {
	processors.Add("printer", func() telex.Processor {
		return &Printer{
			serializer: influx.NewSerializer(),
		}
	})
}
