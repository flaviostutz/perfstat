package stats

import (
	"time"

	"github.com/flaviostutz/signalutils"
)

func TimeLoadPerc(ts *signalutils.Timeseries, loadTime time.Duration) (float64, bool) {
	v1, ok := ts.Get(time.Now().Add(-loadTime))
	if !ok {
		return -1, false
	}
	v2, ok := ts.Last()
	if !ok {
		return -1, false
	}
	//percent of time with load
	td := v2.Time.Sub(v1.Time).Seconds()
	vd := v2.Value - v1.Value
	return vd / td, true
}

func ValuesAvg(ts *signalutils.Timeseries, timeSpan time.Duration) (float64, bool) {
	v1, ok := ts.Get(time.Now().Add(-timeSpan))
	if !ok {
		return -1, false
	}
	v2, ok := ts.Last()
	if !ok {
		return -1, false
	}
	_, v1p, ok := ts.Pos(v1.Time)
	if !ok {
		return 0, false
	}
	v2p, _, ok := ts.Pos(v2.Time)
	if !ok {
		return 0, false
	}

	//average over period
	sum := 0.0
	for i := v1p; i < v2p; i++ {
		sum = sum + ts.Values[i].Value
	}
	return sum / float64(v1p+v2p), true
}
