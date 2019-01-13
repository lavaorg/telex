package inputs

import "github.com/lavaorg/telex"

type Creator func() telex.Input

var Inputs = map[string]Creator{}

func Add(name string, creator Creator) {
	Inputs[name] = creator
}
