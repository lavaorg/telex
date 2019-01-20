package models

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/testutil"

	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	a := &TestAggregator{}
	ra := NewRunningAggregator(a, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			NamePass: []string{"*"},
		},
		Period: time.Millisecond * 500,
	})
	require.NoError(t, ra.Config.Filter.Compile())
	acc := testutil.Accumulator{}

	now := time.Now()
	ra.SetPeriodStart(now)

	m := testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		time.Now().Add(time.Millisecond*150),
		telex.Untyped)
	require.False(t, ra.Add(m))
	ra.Push(&acc)

	require.Equal(t, 1, len(acc.Metrics))
	require.Equal(t, int64(101), acc.Metrics[0].Fields["sum"])
}

func TestAddMetricsOutsideCurrentPeriod(t *testing.T) {
	a := &TestAggregator{}
	ra := NewRunningAggregator(a, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			NamePass: []string{"*"},
		},
		Period: time.Millisecond * 500,
	})
	require.NoError(t, ra.Config.Filter.Compile())
	acc := testutil.Accumulator{}
	now := time.Now()
	ra.SetPeriodStart(now)

	m := testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now.Add(-time.Hour),
		telex.Untyped,
	)
	require.False(t, ra.Add(m))

	// metric after current period
	m = testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now.Add(time.Hour),
		telex.Untyped,
	)
	require.False(t, ra.Add(m))

	// "now" metric
	m = testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		time.Now().Add(time.Millisecond*50),
		telex.Untyped)
	require.False(t, ra.Add(m))

	ra.Push(&acc)
	require.Equal(t, 1, len(acc.Metrics))
	require.Equal(t, int64(202), acc.Metrics[0].Fields["sum"])
}

func TestAddAndPushOnePeriod(t *testing.T) {
	a := &TestAggregator{}
	ra := NewRunningAggregator(a, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			NamePass: []string{"*"},
		},
		Period: time.Millisecond * 500,
	})
	require.NoError(t, ra.Config.Filter.Compile())
	acc := testutil.Accumulator{}

	now := time.Now()
	ra.SetPeriodStart(now)

	m := testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		time.Now().Add(time.Millisecond*100),
		telex.Untyped)
	require.False(t, ra.Add(m))

	ra.Push(&acc)

	acc.AssertContainsFields(t, "TestMetric", map[string]interface{}{"sum": int64(101)})
}

func TestAddDropOriginal(t *testing.T) {
	ra := NewRunningAggregator(&TestAggregator{}, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			NamePass: []string{"RI*"},
		},
		DropOriginal: true,
	})
	require.NoError(t, ra.Config.Filter.Compile())

	now := time.Now()
	ra.SetPeriodStart(now)

	m := testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telex.Untyped)
	require.True(t, ra.Add(m))

	// this metric name doesn't match the filter, so Add will return false
	m2 := testutil.MustMetric("foobar",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telex.Untyped)
	require.False(t, ra.Add(m2))
}

type TestAggregator struct {
	sum int64
}

func (t *TestAggregator) Reset() {
	atomic.StoreInt64(&t.sum, 0)
}

func (t *TestAggregator) Push(acc telex.Accumulator) {
	acc.AddFields("TestMetric",
		map[string]interface{}{"sum": t.sum},
		map[string]string{},
	)
}

func (t *TestAggregator) Add(in telex.Metric) {
	for _, v := range in.Fields() {
		if vi, ok := v.(int64); ok {
			atomic.AddInt64(&t.sum, vi)
		}
	}
}
