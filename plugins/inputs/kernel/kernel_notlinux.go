// +build !linux

package kernel

import (
	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/plugins/inputs"
)

type Kernel struct {
}

func (k *Kernel) Description() string {
	return "Get kernel statistics from /proc/stat"
}

func (k *Kernel) SampleConfig() string { return "" }

func (k *Kernel) Gather(acc telex.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("kernel", func() telex.Input {
		return &Kernel{}
	})
}
