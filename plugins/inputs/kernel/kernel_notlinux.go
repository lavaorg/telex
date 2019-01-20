// +build !linux

package kernel

import (
	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/plugins/inputs"
)

type Kernel struct {
}

func (k *Kernel) Gather(acc telex.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("kernel", func() telex.Input {
		return &Kernel{}
	})
}
