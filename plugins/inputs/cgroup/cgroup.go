package cgroup

import (
	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/plugins/inputs"
)

type CGroup struct {
	Paths []string `toml:"paths"`
	Files []string `toml:"files"`
}

func init() {
	inputs.Add("cgroup", func() telex.Input { return &CGroup{} })
}
