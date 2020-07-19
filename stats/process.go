package stats

import (
	"sort"
	"time"

	"github.com/flaviostutz/signalutils"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
	"github.com/sirupsen/logrus"
)

type ProcessStats struct {
	Processes          map[int32]*ProcessMetrics
	timeseriesMaxSpan  time.Duration
	ioLoadRateTimeSpan time.Duration
	cpuLoadTimeSpan    time.Duration
	worker             *signalutils.Worker
	lastCleanupTime    time.Time
}

type NetIOCounters struct {
	InterfaceName string
	BytesSent     signalutils.TimeseriesCounterRate
	BytesRecv     signalutils.TimeseriesCounterRate
	PacketsSent   signalutils.TimeseriesCounterRate
	PacketsRecv   signalutils.TimeseriesCounterRate
	ErrIn         signalutils.TimeseriesCounterRate
	ErrOut        signalutils.TimeseriesCounterRate
}

type IOCounters struct {
	ReadCount  signalutils.TimeseriesCounterRate
	WriteCount signalutils.TimeseriesCounterRate
	ReadBytes  signalutils.TimeseriesCounterRate
	WriteBytes signalutils.TimeseriesCounterRate
}

type ProcessMetrics struct {
	Pid                int32
	Name               string
	Cmdline            string
	LastSeen           time.Time
	CPUTimes           *CPUTimes
	Connections        signalutils.Timeseries
	TotalNetIOCounters *NetIOCounters
	NetIOCounters      map[string]*NetIOCounters
	IOCounters         *IOCounters
	MemoryPercent      signalutils.Timeseries
	MemoryTotal        signalutils.Timeseries
	FD                 signalutils.Timeseries
	OpenFiles          signalutils.Timeseries
}

func NewProcessStats(timeseriesMaxSpan time.Duration, ioLoadRateTimeSpan time.Duration, cpuLoadTimeSpan time.Duration, sampleFreq float64) *ProcessStats {
	logrus.Tracef("Process Stats: initializing...")
	ps := &ProcessStats{
		Processes:          make(map[int32]*ProcessMetrics),
		ioLoadRateTimeSpan: ioLoadRateTimeSpan,
		cpuLoadTimeSpan:    cpuLoadTimeSpan,
		timeseriesMaxSpan:  timeseriesMaxSpan,
	}
	signalutils.StartWorker("process", ps.processStep, sampleFreq/2, sampleFreq, true)
	logrus.Debugf("Process Stats: running")
	return ps
}

func (ps *ProcessStats) processStep() error {
	//cleanup old processes to avoid memory leaks
	if time.Now().Sub(ps.lastCleanupTime) > 1*time.Hour {
		logrus.Debugf("Performing old processes cleanup...")
		removePids := make([]int32, 0)
		for pid, p := range ps.Processes {
			if time.Now().Sub(p.LastSeen) > 10*time.Minute {
				removePids = append(removePids, pid)
			}
		}
		for _, pi := range removePids {
			delete(ps.Processes, pi)
		}
		ps.lastCleanupTime = time.Now()
	}

	//stats per process
	processes, err := process.Processes()
	if err != nil {
		return err
	}

	for _, p := range processes {
		proc, ok := ps.Processes[p.Pid]
		if !ok {
			//initialize process counter
			proc = &ProcessMetrics{}

			proc.Pid = p.Pid
			proc.Name, err = p.Name()
			if err != nil {
				return err
			}
			proc.Cmdline, err = p.Cmdline()
			if err != nil {
				return err
			}

			proc.CPUTimes = &CPUTimes{}
			proc.CPUTimes.IOWait = signalutils.NewTimeseries(ps.timeseriesMaxSpan)
			proc.CPUTimes.Idle = signalutils.NewTimeseries(ps.timeseriesMaxSpan)
			proc.CPUTimes.Steal = signalutils.NewTimeseries(ps.timeseriesMaxSpan)
			proc.CPUTimes.System = signalutils.NewTimeseries(ps.timeseriesMaxSpan)
			proc.CPUTimes.User = signalutils.NewTimeseries(ps.timeseriesMaxSpan)

			proc.IOCounters = &IOCounters{}
			proc.IOCounters.ReadBytes = signalutils.NewTimeseriesCounterRate(ps.timeseriesMaxSpan)
			proc.IOCounters.ReadCount = signalutils.NewTimeseriesCounterRate(ps.timeseriesMaxSpan)
			proc.IOCounters.WriteBytes = signalutils.NewTimeseriesCounterRate(ps.timeseriesMaxSpan)
			proc.IOCounters.WriteCount = signalutils.NewTimeseriesCounterRate(ps.timeseriesMaxSpan)

			proc.TotalNetIOCounters = &NetIOCounters{}
			proc.TotalNetIOCounters.BytesRecv = signalutils.NewTimeseriesCounterRate(ps.timeseriesMaxSpan)
			proc.TotalNetIOCounters.BytesSent = signalutils.NewTimeseriesCounterRate(ps.timeseriesMaxSpan)
			proc.TotalNetIOCounters.PacketsRecv = signalutils.NewTimeseriesCounterRate(ps.timeseriesMaxSpan)
			proc.TotalNetIOCounters.PacketsSent = signalutils.NewTimeseriesCounterRate(ps.timeseriesMaxSpan)
			proc.TotalNetIOCounters.ErrIn = signalutils.NewTimeseriesCounterRate(ps.timeseriesMaxSpan)
			proc.TotalNetIOCounters.ErrOut = signalutils.NewTimeseriesCounterRate(ps.timeseriesMaxSpan)
			proc.NetIOCounters = make(map[string]*NetIOCounters)

			proc.Connections = signalutils.NewTimeseries(ps.timeseriesMaxSpan)
			proc.FD = signalutils.NewTimeseries(ps.timeseriesMaxSpan)
			proc.OpenFiles = signalutils.NewTimeseries(ps.timeseriesMaxSpan)
			proc.MemoryPercent = signalutils.NewTimeseries(ps.timeseriesMaxSpan)
			proc.MemoryTotal = signalutils.NewTimeseries(ps.timeseriesMaxSpan)

			ps.Processes[p.Pid] = proc
		}
		addProcessStats(p, proc, ps.timeseriesMaxSpan)
	}

	return nil
}

func addProcessStats(p *process.Process, proc *ProcessMetrics, timeseriesMaxSpan time.Duration) {
	var err error
	proc.LastSeen = time.Now()

	//cpu usage
	timestats, err := p.Times()
	if err != nil {
		logrus.Warnf("Error getting process CPUTimes for pid=%d; err=%s", p.Pid, err)
	} else {
		proc.CPUTimes.IOWait.Add(timestats.Iowait)
		proc.CPUTimes.Idle.Add(timestats.Idle)
		proc.CPUTimes.Steal.Add(timestats.Steal)
		proc.CPUTimes.System.Add(timestats.System)
		proc.CPUTimes.User.Add(timestats.User)
	}

	//network connection count
	connstats, err := p.Connections()
	if err != nil {
		logrus.Warnf("Error getting process Connections for pid=%d; err=%s", p.Pid, err)
	} else {
		proc.Connections.Add(float64(len(connstats)))
	}

	//network io overall
	netiostats, err := p.NetIOCounters(false)
	if err != nil {
		logrus.Warnf("Error getting process overall NETIOCounters for pid=%d; err=%s", p.Pid, err)
	} else {
		n := &netiostats[0]
		addNetIOCounters(n, proc.TotalNetIOCounters)
	}

	//network io per interface
	netiostats, err = p.NetIOCounters(true)
	if err != nil {
		logrus.Warnf("Error getting process NETIOCounters per nic for pid=%d; err=%s", p.Pid, err)
	}
	for _, niostat := range netiostats {
		nc, ok := proc.NetIOCounters[niostat.Name]
		if !ok {
			nc = &NetIOCounters{}
			nc.InterfaceName = niostat.Name
			nc.BytesRecv = signalutils.NewTimeseriesCounterRate(timeseriesMaxSpan)
			nc.BytesSent = signalutils.NewTimeseriesCounterRate(timeseriesMaxSpan)
			nc.ErrIn = signalutils.NewTimeseriesCounterRate(timeseriesMaxSpan)
			nc.ErrOut = signalutils.NewTimeseriesCounterRate(timeseriesMaxSpan)
			nc.PacketsRecv = signalutils.NewTimeseriesCounterRate(timeseriesMaxSpan)
			nc.PacketsSent = signalutils.NewTimeseriesCounterRate(timeseriesMaxSpan)
			proc.NetIOCounters[niostat.Name] = nc
		}
		addNetIOCounters(&niostat, nc)
	}

	//io counters
	cs, err := p.IOCounters()
	if err != nil {
		logrus.Warnf("Error getting process IOCounters for pid=%d; err=%s", p.Pid, err)
	} else {
		ioc := proc.IOCounters
		ioc.ReadBytes.Set(float64(cs.ReadBytes))
		ioc.ReadCount.Set(float64(cs.ReadCount))
		ioc.WriteBytes.Set(float64(cs.WriteBytes))
		ioc.WriteCount.Set(float64(cs.WriteCount))
	}

	//ram memory
	mp, err := p.MemoryPercent()
	if err != nil {
		logrus.Warnf("Error getting process MemoryPercent for pid=%d; err=%s", p.Pid, err)
	} else {
		proc.MemoryPercent.Add(float64(mp))
	}

	mi, err := p.MemoryInfo()
	if err != nil {
		logrus.Warnf("Error getting process MemoryInfo for pid=%d; err=%s", p.Pid, err)
	} else {
		proc.MemoryTotal.Add(float64(mi.RSS))
	}

	//file descriptors
	fd, err := p.NumFDs()
	if err != nil {
		logrus.Warnf("Error getting process NumFDs for pid=%d; err=%s", p.Pid, err)
	} else {
		proc.FD.Add(float64(fd))
	}

	//open files
	of, err := p.OpenFiles()
	if err != nil {
		logrus.Warnf("Error getting process OpenFiles for pid=%d; err=%s", p.Pid, err)
	} else {
		proc.OpenFiles.Add(float64(len(of)))
	}
}

func addNetIOCounters(n *net.IOCountersStat, nioc *NetIOCounters) {
	nioc.InterfaceName = n.Name
	nioc.BytesRecv.Set(float64(n.BytesRecv))
	nioc.BytesSent.Set(float64(n.BytesSent))
	nioc.ErrIn.Set(float64(n.Errin + n.Dropin))
	nioc.ErrOut.Set(float64(n.Errout + n.Dropout))
	nioc.PacketsRecv.Set(float64(n.PacketsRecv))
	nioc.PacketsSent.Set(float64(n.PacketsSent))
}

func (ps *ProcessStats) Stop() {
	ps.worker.Stop()
}

func (p *ProcessStats) TopCPULoad() []*ProcessMetrics {
	pa := p.processesArray()
	sort.Slice(pa, func(i, j int) bool {
		pi := pa[i]
		pj := pa[j]

		sysi, _ := TimeLoadPerc(&pi.CPUTimes.System, p.cpuLoadTimeSpan)
		usri, _ := TimeLoadPerc(&pi.CPUTimes.User, p.cpuLoadTimeSpan)

		sysj, _ := TimeLoadPerc(&pj.CPUTimes.System, p.cpuLoadTimeSpan)
		usrj, _ := TimeLoadPerc(&pj.CPUTimes.User, p.cpuLoadTimeSpan)

		return (sysj + usrj) < (sysi + usri)
	})
	return pa
}

func (p *ProcessStats) TopCPUIOWait() []*ProcessMetrics {
	pa := p.processesArray()
	sort.Slice(pa, func(i, j int) bool {
		pi := pa[i]
		pj := pa[j]

		wi, _ := TimeLoadPerc(&pi.CPUTimes.IOWait, p.cpuLoadTimeSpan)
		wj, _ := TimeLoadPerc(&pj.CPUTimes.IOWait, p.cpuLoadTimeSpan)

		return wj < wi
	})
	return pa
}

func (p *ProcessStats) TopIOByteRate(read bool) []*ProcessMetrics {
	pa := p.processesArray()
	sort.Slice(pa, func(i, j int) bool {
		pi := pa[i]
		pj := pa[j]

		if read {
			ri, _ := pi.IOCounters.ReadBytes.Rate(p.ioLoadRateTimeSpan)
			rj, _ := pj.IOCounters.ReadBytes.Rate(p.ioLoadRateTimeSpan)
			return rj < ri
		}

		wi, _ := pi.IOCounters.WriteBytes.Rate(p.ioLoadRateTimeSpan)
		wj, _ := pj.IOCounters.WriteBytes.Rate(p.ioLoadRateTimeSpan)
		return wj < wi

	})
	return pa
}

func (p *ProcessStats) TopIOOpRate(read bool) []*ProcessMetrics {
	pa := p.processesArray()
	sort.Slice(pa, func(i, j int) bool {
		pi := pa[i]
		pj := pa[j]

		if read {
			ri, _ := pi.IOCounters.ReadCount.Rate(p.ioLoadRateTimeSpan)
			rj, _ := pj.IOCounters.ReadCount.Rate(p.ioLoadRateTimeSpan)
			return rj < ri
		}

		wi, _ := pi.IOCounters.WriteCount.Rate(p.ioLoadRateTimeSpan)
		wj, _ := pj.IOCounters.WriteCount.Rate(p.ioLoadRateTimeSpan)
		return wj < wi

	})
	return pa
}

func (p *ProcessStats) TopMemUsed() []*ProcessMetrics {
	pa := p.processesArray()
	to := time.Now()
	from := to.Add(-p.ioLoadRateTimeSpan)
	sort.Slice(pa, func(i, j int) bool {
		pi := pa[i]
		pj := pa[j]

		vi, _ := pi.MemoryTotal.Avg(from, to)
		vj, _ := pj.MemoryTotal.Avg(from, to)

		return vj < vi
	})
	return pa
}

func (p *ProcessStats) TopFD() []*ProcessMetrics {
	pa := p.processesArray()
	to := time.Now()
	from := to.Add(-p.ioLoadRateTimeSpan)
	sort.Slice(pa, func(i, j int) bool {
		pi := pa[i]
		pj := pa[j]

		vi, _ := pi.FD.Avg(from, to)
		vj, _ := pj.FD.Avg(from, to)

		return vj < vi
	})
	return pa
}

func (p *ProcessStats) TopNetByteRate(recv bool) []*ProcessMetrics {
	pa := p.processesArray()
	sort.Slice(pa, func(i, j int) bool {
		pi := pa[i]
		pj := pa[j]

		if recv {
			ri, _ := pi.TotalNetIOCounters.BytesRecv.Rate(p.ioLoadRateTimeSpan)
			rj, _ := pj.TotalNetIOCounters.BytesRecv.Rate(p.ioLoadRateTimeSpan)
			return rj < ri
		}

		si, _ := pi.TotalNetIOCounters.BytesSent.Rate(p.ioLoadRateTimeSpan)
		sj, _ := pj.TotalNetIOCounters.BytesSent.Rate(p.ioLoadRateTimeSpan)
		return sj < si
	})
	return pa
}

func (p *ProcessStats) TopNetErrRate(in bool) []*ProcessMetrics {
	pa := p.processesArray()
	sort.Slice(pa, func(i, j int) bool {
		pi := pa[i]
		pj := pa[j]

		if in {
			ri, _ := pi.TotalNetIOCounters.ErrIn.Rate(p.ioLoadRateTimeSpan)
			rj, _ := pj.TotalNetIOCounters.ErrIn.Rate(p.ioLoadRateTimeSpan)
			return rj < ri
		}

		si, _ := pi.TotalNetIOCounters.ErrOut.Rate(p.ioLoadRateTimeSpan)
		sj, _ := pj.TotalNetIOCounters.ErrOut.Rate(p.ioLoadRateTimeSpan)
		return sj < si
	})
	return pa
}

func (p *ProcessStats) TopNetConnCount() []*ProcessMetrics {
	pa := p.processesArray()
	to := time.Now()
	from := to.Add(-p.ioLoadRateTimeSpan)
	sort.Slice(pa, func(i, j int) bool {
		pi := pa[i]
		pj := pa[j]

		si, _ := pi.Connections.Avg(from, to)
		sj, _ := pj.Connections.Avg(from, to)
		return sj < si
	})
	return pa
}

func (p *ProcessStats) processesArray() []*ProcessMetrics {
	pa := make([]*ProcessMetrics, 0)
	for _, v := range p.Processes {
		pa = append(pa, v)
	}
	return pa
}
