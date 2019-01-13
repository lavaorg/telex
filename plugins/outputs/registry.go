package outputs

import (
	"github.com/lavaorg/telex"
)

type Creator func() telex.Output

var Outputs = map[string]Creator{}

func Add(name string, creator Creator) {
	Outputs[name] = creator
}
