// +build !linux

package zfs

import (
	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/plugins/inputs"
)

func (z *Zfs) Gather(acc telex.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("zfs", func() telex.Input {
		return &Zfs{}
	})
}
