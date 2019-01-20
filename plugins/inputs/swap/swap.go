package swap

import (
	"fmt"

	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/plugins/inputs"
	"github.com/lavaorg/telex/plugins/inputs/system"
)

type SwapStats struct {
	ps system.PS
}

func (s *SwapStats) Gather(acc telex.Accumulator) error {
	swap, err := s.ps.SwapStat()
	if err != nil {
		return fmt.Errorf("error getting swap memory info: %s", err)
	}

	fieldsG := map[string]interface{}{
		"total":        swap.Total,
		"used":         swap.Used,
		"free":         swap.Free,
		"used_percent": swap.UsedPercent,
	}
	fieldsC := map[string]interface{}{
		"in":  swap.Sin,
		"out": swap.Sout,
	}
	acc.AddGauge("swap", fieldsG, nil)
	acc.AddCounter("swap", fieldsC, nil)

	return nil
}

func init() {
	ps := system.NewSystemPS()
	inputs.Add("swap", func() telex.Input {
		return &SwapStats{ps: ps}
	})
}
