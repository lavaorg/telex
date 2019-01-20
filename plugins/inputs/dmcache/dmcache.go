package dmcache

import (
	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/plugins/inputs"
)

type DMCache struct {
	PerDevice        bool `toml:"per_device"`
	getCurrentStatus func() ([]string, error)
}


func init() {
	inputs.Add("dmcache", func() telex.Input {
		return &DMCache{
			PerDevice:        true,
			getCurrentStatus: dmSetupStatus,
		}
	})
}
