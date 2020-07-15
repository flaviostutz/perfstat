package detectors

import (
	"time"

	"github.com/flaviostutz/signalutils"
	"github.com/shirou/gopsutil/cpu"
	"github.com/sirupsen/logrus"
)

var (
	cpuStats             = &CPUStats{}
	cpuTimeseriesMaxSpan = 10 * time.Minute
)

type CPUTimes struct {
	Idle   signalutils.Timeseries
	System signalutils.Timeseries
	User   signalutils.Timeseries
	IOWait signalutils.Timeseries
	Steal  signalutils.Timeseries
}

type CPUStats struct {
	Total CPUTimes
	CPU   []CPUTimes
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
	return (v2.Value - v1.Value) / (v2.Time.Sub(v1.Time).Seconds()), true
}

func StartCPUWatcher() {
	logrus.Debugf("CPU Watcher: initializing...")

	nrcpu, err := cpu.Counts(true)
	if err != nil {
		logrus.Warningf("Cannot initilize cpu watcher. err=%s", err)
	}

	cpuStats.CPU = make([]CPUTimes, 0)
	for i := 0; i < nrcpu; i++ {
		cpuStats.CPU = append(cpuStats.CPU, newCPUTimes(cpuTimeseriesMaxSpan))
	}
	cpuStats.Total = newCPUTimes(cpuTimeseriesMaxSpan)

	StartWorker("cpu", cpuStep, 1, true)
	logrus.Debugf("CPU Watcher: running")
}

func newCPUTimes(tsDuration time.Duration) CPUTimes {
	return CPUTimes{
		Idle:   signalutils.NewTimeseries(tsDuration),
		System: signalutils.NewTimeseries(tsDuration),
		User:   signalutils.NewTimeseries(tsDuration),
		IOWait: signalutils.NewTimeseries(tsDuration),
		Steal:  signalutils.NewTimeseries(tsDuration),
	}
}

func addCPUStats(cpu *cpu.TimesStat, cpuTimes *CPUTimes) {
	cpuTimes.Idle.AddSample(cpu.Idle)
	cpuTimes.System.AddSample(cpu.System)
	cpuTimes.User.AddSample(cpu.User)
	cpuTimes.IOWait.AddSample(cpu.Iowait)
	cpuTimes.Steal.AddSample(cpu.Steal)
	logrus.Debugf("cpustats=%s", cpu.String())
}

func cpuStep() error {
	//overall load
	cpus, err := cpu.Times(false)
	if err != nil {
		return err
	}
	addCPUStats(&cpus[0], &cpuStats.Total)

	//load per CPU
	cpus, err = cpu.Times(true)
	if err != nil {
		return err
	}
	for i, cs := range cpuStats.CPU {
		addCPUStats(&cpus[i], &cs)
	}

	return nil
}
