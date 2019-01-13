package metric

import (
	"log"
	"runtime"
	"sync/atomic"

	"github.com/lavaorg/telex"
)

// NotifyFunc is called when a tracking metric is done being processed with
// the tracking information.
type NotifyFunc = func(track telex.DeliveryInfo)

// WithTracking adds tracking to the metric and registers the notify function
// to be called when processing is complete.
func WithTracking(metric telex.Metric, fn NotifyFunc) (telex.Metric, telex.TrackingID) {
	return newTrackingMetric(metric, fn)
}

// WithBatchTracking adds tracking to the metrics and registers the notify
// function to be called when processing is complete.
func WithGroupTracking(metric []telex.Metric, fn NotifyFunc) ([]telex.Metric, telex.TrackingID) {
	return newTrackingMetricGroup(metric, fn)
}

func EnableDebugFinalizer() {
	finalizer = debugFinalizer
}

var (
	lastID    uint64
	finalizer func(*trackingData)
)

func newTrackingID() telex.TrackingID {
	atomic.AddUint64(&lastID, 1)
	return telex.TrackingID(lastID)
}

func debugFinalizer(d *trackingData) {
	rc := atomic.LoadInt32(&d.rc)
	if rc != 0 {
		log.Fatalf("E! [agent] metric collected with non-zero reference count rc: %d", rc)
	}
}

type trackingData struct {
	id          telex.TrackingID
	rc          int32
	acceptCount int32
	rejectCount int32
	notifyFunc  NotifyFunc
}

func (d *trackingData) incr() {
	atomic.AddInt32(&d.rc, 1)
}

func (d *trackingData) decr() int32 {
	return atomic.AddInt32(&d.rc, -1)
}

func (d *trackingData) accept() {
	atomic.AddInt32(&d.acceptCount, 1)
}

func (d *trackingData) reject() {
	atomic.AddInt32(&d.rejectCount, 1)
}

func (d *trackingData) notify() {
	d.notifyFunc(
		&deliveryInfo{
			id:       d.id,
			accepted: int(d.acceptCount),
			rejected: int(d.rejectCount),
		},
	)
}

type trackingMetric struct {
	telex.Metric
	d *trackingData
}

func newTrackingMetric(metric telex.Metric, fn NotifyFunc) (telex.Metric, telex.TrackingID) {
	m := &trackingMetric{
		Metric: metric,
		d: &trackingData{
			id:          newTrackingID(),
			rc:          1,
			acceptCount: 0,
			rejectCount: 0,
			notifyFunc:  fn,
		},
	}

	if finalizer != nil {
		runtime.SetFinalizer(m.d, finalizer)
	}
	return m, m.d.id
}

func newTrackingMetricGroup(group []telex.Metric, fn NotifyFunc) ([]telex.Metric, telex.TrackingID) {
	d := &trackingData{
		id:          newTrackingID(),
		rc:          0,
		acceptCount: 0,
		rejectCount: 0,
		notifyFunc:  fn,
	}

	for i, m := range group {
		d.incr()
		dm := &trackingMetric{
			Metric: m,
			d:      d,
		}
		group[i] = dm

	}
	if finalizer != nil {
		runtime.SetFinalizer(d, finalizer)
	}

	if len(group) == 0 {
		d.notify()
	}

	return group, d.id
}

func (m *trackingMetric) Copy() telex.Metric {
	m.d.incr()
	return &trackingMetric{
		Metric: m.Metric.Copy(),
		d:      m.d,
	}
}

func (m *trackingMetric) Accept() {
	m.d.accept()
	m.decr()
}

func (m *trackingMetric) Reject() {
	m.d.reject()
	m.decr()
}

func (m *trackingMetric) Drop() {
	m.decr()
}

func (m *trackingMetric) decr() {
	v := m.d.decr()
	if v < 0 {
		panic("negative refcount")
	}

	if v == 0 {
		m.d.notify()
	}
}

type deliveryInfo struct {
	id       telex.TrackingID
	accepted int
	rejected int
}

func (r *deliveryInfo) ID() telex.TrackingID {
	return r.id
}

func (r *deliveryInfo) Delivered() bool {
	return r.rejected == 0
}
