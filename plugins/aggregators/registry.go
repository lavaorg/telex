package aggregators

import "github.com/lavaorg/telex"

type Creator func() telex.Aggregator

var Aggregators = map[string]Creator{}

func Add(name string, creator Creator) {
	Aggregators[name] = creator
}
