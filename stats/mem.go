package stats

import (
	"time"

	"github.com/flaviostutz/signalutils"
	"github.com/shirou/gopsutil/mem"
	"github.com/sirupsen/logrus"
)

type MemStats struct {
	Total     uint64
	Available signalutils.Timeseries
	Used      signalutils.Timeseries
	Free      signalutils.Timeseries
	SwapIn    signalutils.TimeseriesCounterRate
	SwapOut   signalutils.TimeseriesCounterRate
	SwapTotal uint64
	SwapUsed  signalutils.Timeseries
	SwapFree  signalutils.Timeseries
	worker    *signalutils.Worker
}

func NewMemStats(timeseriesMaxSpan time.Duration, sampleFreq float64) *MemStats {
	logrus.Tracef("Mem Stats: initializing...")

	mt := &MemStats{}
	mt.Total = 0.0
	mt.Available = signalutils.Timeseries{}
	mt.Used = signalutils.Timeseries{}
	mt.Free = signalutils.Timeseries{}
	mt.SwapIn = signalutils.TimeseriesCounterRate{}
	mt.SwapOut = signalutils.TimeseriesCounterRate{}
	mt.SwapTotal = 0.0
	mt.SwapUsed = signalutils.Timeseries{}

	mt.worker = signalutils.StartWorker("mem", mt.memStep, sampleFreq/2, sampleFreq, true)
	logrus.Debugf("Mem Stats: running")

	return mt
}

func (m *MemStats) Stop() {
	m.worker.Stop()
}

func (m *MemStats) memStep() error {

	ms, err := mem.VirtualMemory()
	if err != nil {
		logrus.Warningf("Cannot initilize mem stats. err=%s", err)
	}

	ss, err := mem.SwapMemory()
	if err != nil {
		logrus.Warningf("Cannot initilize swap stats. err=%s", err)
	}

	m.Total = ms.Total
	m.Used.Add(float64(ms.Used))
	m.Available.Add(float64(ms.Available))
	m.Free.Add(float64(ms.Free))
	m.SwapTotal = ss.Total
	m.SwapUsed.Add(float64(ss.Used))
	m.SwapFree.Add(float64(ss.Free))
	m.SwapIn.Set(float64(ss.Sin))
	m.SwapOut.Set(float64(ss.Sout))

	return nil
}
