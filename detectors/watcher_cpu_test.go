package detectors

import (
	"testing"
	"time"

	"github.com/flaviostutz/signalutils"
	"github.com/stretchr/testify/assert"
)

func TestCPUWatcherTotal(t *testing.T) {
	// logrus.SetLevel(logrus.DebugLevel)
	StartCPUWatcher()
	time.Sleep(5 * time.Second)
	tc, ok := CPUAvgPerc(&cpuStats.Total.Idle, 3*time.Second)
	// fmt.Printf(">>>>> %f\n", tc)
	assert.True(t, ok)
	assert.GreaterOrEqualf(t, tc, 0.3, "")
	// assert.LessOrEqualf(t, tc, 1.0, "")
}

func TestCPUWatcherPerCPU(t *testing.T) {
	// logrus.SetLevel(logrus.DebugLevel)
	StartCPUWatcher()
	time.Sleep(5 * time.Second)
	// tc1, ok := CPUAvgPerc(&cpuStats.CPU[0].Idle, 3*time.Second)
	tc2, ok := CPUAvgPerc(&cpuStats.CPU[0].User, 3*time.Second)
	// tc3, ok := CPUAvgPerc(&cpuStats.CPU[0].System, 3*time.Second)
	// fmt.Printf(">>>>> IDLE0=%f USER0=%f SYSTEM=%f\n", tc1, tc2, tc3)
	assert.True(t, ok)
	assert.LessOrEqualf(t, tc2, 0.4, "")
	// assert.LessOrEqualf(t, tc, 1.0, "")
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
