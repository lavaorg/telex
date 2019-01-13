package processors

import "github.com/lavaorg/telex"

type Creator func() telex.Processor

var Processors = map[string]Creator{}

func Add(name string, creator Creator) {
	Processors[name] = creator
}
