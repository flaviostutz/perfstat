package detectors

import (
	"testing"
	"time"

	"github.com/flaviostutz/signalutils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCPUWatcher(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	StartCPUWatcher()
	time.Sleep(5 * time.Second)
	tc, ok := CPUAvgPerc(&cpuStats.Total.Idle, 1*time.Second)
	assert.True(t, ok)
	assert.GreaterOrEqualf(t, tc, 0.3, "")
	assert.LessOrEqualf(t, tc, 1.0, "")
}

func TestCPUAvgPerc(t *testing.T) {
	ts := signalutils.NewTimeseries(10 * time.Second)
	ts.AddSample(10)
	time.Sleep(1 * time.Second)
	ts.AddSample(10.6)
	v, ok := CPUAvgPerc(&ts, 500*time.Millisecond)
	assert.True(t, ok)
	assert.InDeltaf(t, 0.6, v, 0.01, "")
}
