// +build !linux

package dmcache

import (
	"github.com/lavaorg/telex"
)

func (c *DMCache) Gather(acc telex.Accumulator) error {
	return nil
}

func dmSetupStatus() ([]string, error) {
	return []string{}, nil
}
