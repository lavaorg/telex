package temp

import (
	"fmt"
	"strings"

	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/plugins/inputs"
	"github.com/lavaorg/telex/plugins/inputs/system"
)

type Temperature struct {
	ps system.PS
}

func (t *Temperature) Gather(acc telex.Accumulator) error {
	temps, err := t.ps.Temperature()
	if err != nil {
		if strings.Contains(err.Error(), "not implemented yet") {
			return fmt.Errorf("plugin is not supported on this platform: %v", err)
		}
		return fmt.Errorf("error getting temperatures info: %s", err)
	}
	for _, temp := range temps {
		tags := map[string]string{
			"sensor": temp.SensorKey,
		}
		fields := map[string]interface{}{
			"temp": temp.Temperature,
		}
		acc.AddFields("temp", fields, tags)
	}
	return nil
}

func init() {
	inputs.Add("temp", func() telex.Input {
		return &Temperature{ps: system.NewSystemPS()}
	})
}
