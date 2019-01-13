package testutil

import (
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/metric"
)

type metricDiff struct {
	Measurement string
	Tags        []*telex.Tag
	Fields      []*telex.Field
	Type        telex.ValueType
	Time        time.Time
}

func newMetricDiff(metric telex.Metric) *metricDiff {
	m := &metricDiff{}
	m.Measurement = metric.Name()

	for _, tag := range metric.TagList() {
		m.Tags = append(m.Tags, tag)
	}
	sort.Slice(m.Tags, func(i, j int) bool {
		return m.Tags[i].Key < m.Tags[j].Key
	})

	for _, field := range metric.FieldList() {
		m.Fields = append(m.Fields, field)
	}
	sort.Slice(m.Fields, func(i, j int) bool {
		return m.Fields[i].Key < m.Fields[j].Key
	})

	m.Type = metric.Type()
	m.Time = metric.Time()
	return m
}

func MetricEqual(expected, actual telex.Metric) bool {
	var lhs, rhs *metricDiff
	if expected != nil {
		lhs = newMetricDiff(expected)
	}
	if actual != nil {
		rhs = newMetricDiff(actual)
	}

	return cmp.Equal(lhs, rhs)
}

func RequireMetricEqual(t *testing.T, expected, actual telex.Metric) {
	t.Helper()

	var lhs, rhs *metricDiff
	if expected != nil {
		lhs = newMetricDiff(expected)
	}
	if actual != nil {
		rhs = newMetricDiff(actual)
	}

	if diff := cmp.Diff(lhs, rhs); diff != "" {
		t.Fatalf("telex.Metric\n--- expected\n+++ actual\n%s", diff)
	}
}

func RequireMetricsEqual(t *testing.T, expected, actual []telex.Metric) {
	t.Helper()

	lhs := make([]*metricDiff, 0, len(expected))
	for _, m := range expected {
		lhs = append(lhs, newMetricDiff(m))
	}
	rhs := make([]*metricDiff, 0, len(actual))
	for _, m := range actual {
		rhs = append(rhs, newMetricDiff(m))
	}
	if diff := cmp.Diff(lhs, rhs); diff != "" {
		t.Fatalf("[]telex.Metric\n--- expected\n+++ actual\n%s", diff)
	}
}

// Metric creates a new metric or panics on error.
func MustMetric(
	name string,
	tags map[string]string,
	fields map[string]interface{},
	tm time.Time,
	tp ...telex.ValueType,
) telex.Metric {
	m, err := metric.New(name, tags, fields, tm, tp...)
	if err != nil {
		panic("MustMetric")
	}
	return m
}
