// +build !linux

package cgroup

import (
	"github.com/lavaorg/telex"
)

func (g *CGroup) Gather(acc telex.Accumulator) error {
	return nil
}
