### Processor Plugins

This section is for developers who want to create a new processor plugin.

### Processor Plugin Guidelines

* A processor must conform to the [telex.Processor][] interface.
* Processors should call `processors.Add` in their `init` function to register
  themselves.  See below for a quick example.
* To be available within Telex itself, plugins must add themselves to the
  `github.com/lavaorg/telex/plugins/processors/all/all.go` file.
* The `SampleConfig` function should return valid toml that describes how the
  processor can be configured. This is include in the output of `telex
  config`.
- The `SampleConfig` function should return valid toml that describes how the
  plugin can be configured. This is included in `telex config`.  Please
  consult the [SampleConfig][] page for the latest style guidelines.
* The `Description` function should say in one line what this processor does.
- Follow the recommended [CodeStyle][].

### Processor Plugin Example

```go
package printer

// printer.go

import (
	"fmt"

	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/plugins/processors"
)

type Printer struct {
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
		fmt.Println(metric.String())
	}
	return in
}

func init() {
	processors.Add("printer", func() telex.Processor {
		return &Printer{}
	})
}
```

[SampleConfig]: https://github.com/lavaorg/telex/wiki/SampleConfig
[CodeStyle]: https://github.com/lavaorg/telex/wiki/CodeStyle
[telex.Processor]: https://godoc.org/github.com/lavaorg/telex#Processor
