package trig

import (
	"math"

	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/plugins/inputs"
)

type Trig struct {
	x         float64
	Amplitude float64
}

func (s *Trig) Gather(acc telex.Accumulator) error {
	sinner := math.Sin((s.x*math.Pi)/5.0) * s.Amplitude
	cosinner := math.Cos((s.x*math.Pi)/5.0) * s.Amplitude

	fields := make(map[string]interface{})
	fields["sine"] = sinner
	fields["cosine"] = cosinner

	tags := make(map[string]string)

	s.x += 1.0
	acc.AddFields("trig", fields, tags)

	return nil
}

func init() {
	inputs.Add("trig", func() telex.Input { return &Trig{x: 0.0} })
}
