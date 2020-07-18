package stats

import (
	"time"

	"github.com/flaviostutz/signalutils"
	"github.com/shirou/gopsutil/cpu"
	"github.com/sirupsen/logrus"
)

type CPUStats struct {
	Total  *CPUTimes
	CPU    []*CPUTimes
	worker *signalutils.Worker
}

type CPUTimes struct {
	Idle   signalutils.Timeseries
	System signalutils.Timeseries
	User   signalutils.Timeseries
	IOWait signalutils.Timeseries
	Steal  signalutils.Timeseries
}

func CPUAvgPerc(ts *signalutils.Timeseries, loadTime time.Duration) (float64, bool) {
	v1, ok := ts.GetValue(time.Now().Add(-loadTime))
	if !ok {
		return -1, false
	}
	v2, ok := ts.GetLastValue()
	if !ok {
		return -1, false
	}
	//percent of time with load
	td := v2.Time.Sub(v1.Time).Seconds()
	vd := v2.Value - v1.Value
	return vd / td, true
}

func NewCPUStats(timeseriesMaxSpan time.Duration, sampleFreq float64) *CPUStats {
	logrus.Tracef("CPU Stats: initializing...")

	nrcpu, err := cpu.Counts(true)
	if err != nil {
		logrus.Warningf("Cannot initilize cpu stats. err=%s", err)
	}

	ct := &CPUStats{}
	ct.CPU = make([]*CPUTimes, 0)
	for i := 0; i < nrcpu; i++ {
		ct.CPU = append(ct.CPU, newCPUTimes(timeseriesMaxSpan))
	}
	ct.Total = newCPUTimes(timeseriesMaxSpan)

	ct.worker = signalutils.StartWorker("cpu", ct.cpuStep, sampleFreq/2, sampleFreq, true)
	logrus.Debugf("CPU Stats: running")

	return ct
}

func (c *CPUStats) Stop() {
	c.worker.Stop()
}

func newCPUTimes(tsDuration time.Duration) *CPUTimes {
	return &CPUTimes{
		Idle:   signalutils.NewTimeseries(tsDuration),
		System: signalutils.NewTimeseries(tsDuration),
		User:   signalutils.NewTimeseries(tsDuration),
		IOWait: signalutils.NewTimeseries(tsDuration),
		Steal:  signalutils.NewTimeseries(tsDuration),
	}
}

func addCPUStats(cpu *cpu.TimesStat, cpuTimes *CPUTimes, cpus float64) {
	cpuTimes.Idle.AddSample(cpu.Idle / cpus)
	cpuTimes.System.AddSample(cpu.System / cpus)
	cpuTimes.User.AddSample(cpu.User / cpus)
	cpuTimes.IOWait.AddSample(cpu.Iowait / cpus)
	cpuTimes.Steal.AddSample(cpu.Steal / cpus)
	// logrus.Debugf("cpustats=%s", cpu.String())
}

func (c *CPUStats) cpuStep() error {
	//overall load
	cpus, err := cpu.Times(false)
	if err != nil {
		return err
	}
	addCPUStats(&cpus[0], c.Total, float64(len(c.CPU)))

	//load per CPU
	cpus, err = cpu.Times(true)
	if err != nil {
		return err
	}
	for i := range c.CPU {
		cs := c.CPU[i]
		addCPUStats(&cpus[i], cs, 1)
	}

	return nil
}
