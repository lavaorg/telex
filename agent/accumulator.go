package agent

import (
	"log"
	"time"

	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/metric"
	"github.com/lavaorg/telex/selfstat"
)

var (
	NErrors = selfstat.Register("agent", "gather_errors", map[string]string{})
)

type MetricMaker interface {
	Name() string
	MakeMetric(metric telex.Metric) telex.Metric
}

type accumulator struct {
	maker     MetricMaker
	metrics   chan<- telex.Metric
	precision time.Duration
}

func NewAccumulator(
	maker MetricMaker,
	metrics chan<- telex.Metric,
) telex.Accumulator {
	acc := accumulator{
		maker:     maker,
		metrics:   metrics,
		precision: time.Nanosecond,
	}
	return &acc
}

func (ac *accumulator) AddFields(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	ac.addFields(measurement, tags, fields, telex.Untyped, t...)
}

func (ac *accumulator) AddGauge(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	ac.addFields(measurement, tags, fields, telex.Gauge, t...)
}

func (ac *accumulator) AddCounter(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	ac.addFields(measurement, tags, fields, telex.Counter, t...)
}

func (ac *accumulator) AddSummary(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	ac.addFields(measurement, tags, fields, telex.Summary, t...)
}

func (ac *accumulator) AddHistogram(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	ac.addFields(measurement, tags, fields, telex.Histogram, t...)
}

func (ac *accumulator) AddMetric(m telex.Metric) {
	m.SetTime(m.Time().Round(ac.precision))
	if m := ac.maker.MakeMetric(m); m != nil {
		ac.metrics <- m
	}
}

func (ac *accumulator) addFields(
	measurement string,
	tags map[string]string,
	fields map[string]interface{},
	tp telex.ValueType,
	t ...time.Time,
) {
	m, err := metric.New(measurement, tags, fields, ac.getTime(t), tp)
	if err != nil {
		return
	}
	if m := ac.maker.MakeMetric(m); m != nil {
		ac.metrics <- m
	}
}

// AddError passes a runtime error to the accumulator.
// The error will be tagged with the plugin name and written to the log.
func (ac *accumulator) AddError(err error) {
	if err == nil {
		return
	}
	NErrors.Incr(1)
	log.Printf("E! [%s]: Error in plugin: %v", ac.maker.Name(), err)
}

func (ac *accumulator) SetPrecision(precision, interval time.Duration) {
	if precision > 0 {
		ac.precision = precision
		return
	}
	switch {
	case interval >= time.Second:
		ac.precision = time.Second
	case interval >= time.Millisecond:
		ac.precision = time.Millisecond
	case interval >= time.Microsecond:
		ac.precision = time.Microsecond
	default:
		ac.precision = time.Nanosecond
	}
}

func (ac *accumulator) getTime(t []time.Time) time.Time {
	var timestamp time.Time
	if len(t) > 0 {
		timestamp = t[0]
	} else {
		timestamp = time.Now()
	}
	return timestamp.Round(ac.precision)
}

func (ac *accumulator) WithTracking(maxTracked int) telex.TrackingAccumulator {
	return &trackingAccumulator{
		Accumulator: ac,
		delivered:   make(chan telex.DeliveryInfo, maxTracked),
	}
}

type trackingAccumulator struct {
	telex.Accumulator
	delivered chan telex.DeliveryInfo
}

func (a *trackingAccumulator) AddTrackingMetric(m telex.Metric) telex.TrackingID {
	dm, id := metric.WithTracking(m, a.onDelivery)
	a.AddMetric(dm)
	return id
}

func (a *trackingAccumulator) AddTrackingMetricGroup(group []telex.Metric) telex.TrackingID {
	db, id := metric.WithGroupTracking(group, a.onDelivery)
	for _, m := range db {
		a.AddMetric(m)
	}
	return id
}

func (a *trackingAccumulator) Delivered() <-chan telex.DeliveryInfo {
	return a.delivered
}

func (a *trackingAccumulator) onDelivery(info telex.DeliveryInfo) {
	select {
	case a.delivered <- info:
	default:
		// This is a programming error in the input.  More items were sent for
		// tracking than space requested.
		panic("channel is full")
	}
}
