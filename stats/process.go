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
	Processes         map[int32]*ProcessCounters
	timeseriesMaxSpan time.Duration
	worker            *signalutils.Worker
	lastCleanupTime   time.Time
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

type ProcessCounters struct {
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

func NewProcessStats(timeseriesMaxSpan time.Duration, sampleFreq float64) *ProcessStats {
	logrus.Tracef("Process Stats: initializing...")
	ps := &ProcessStats{
		Processes:         make(map[int32]*ProcessCounters),
		timeseriesMaxSpan: timeseriesMaxSpan,
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
			proc = &ProcessCounters{}

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

func addProcessStats(p *process.Process, proc *ProcessCounters, timeseriesMaxSpan time.Duration) {
	var err error
	proc.LastSeen = time.Now()

	//cpu usage
	timestats, err := p.Times()
	if err != nil {
		logrus.Warnf("Error getting process CPUTimes for pid=%d; err=%s", p.Pid, err)
	} else {
		proc.CPUTimes.IOWait.AddSample(timestats.Iowait)
		proc.CPUTimes.Idle.AddSample(timestats.Idle)
		proc.CPUTimes.Steal.AddSample(timestats.Steal)
		proc.CPUTimes.System.AddSample(timestats.System)
		proc.CPUTimes.User.AddSample(timestats.User)
	}

	//network connection count
	connstats, err := p.Connections()
	if err != nil {
		logrus.Warnf("Error getting process Connections for pid=%d; err=%s", p.Pid, err)
	} else {
		proc.Connections.AddSample(float64(len(connstats)))
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
		proc.MemoryPercent.AddSample(float64(mp))
	}

	mi, err := p.MemoryInfo()
	if err != nil {
		logrus.Warnf("Error getting process MemoryInfo for pid=%d; err=%s", p.Pid, err)
	} else {
		proc.MemoryTotal.AddSample(float64(mi.RSS))
	}

	//file descriptors
	fd, err := p.NumFDs()
	if err != nil {
		logrus.Warnf("Error getting process NumFDs for pid=%d; err=%s", p.Pid, err)
	} else {
		proc.FD.AddSample(float64(fd))
	}

	//open files
	of, err := p.OpenFiles()
	if err != nil {
		logrus.Warnf("Error getting process OpenFiles for pid=%d; err=%s", p.Pid, err)
	} else {
		proc.OpenFiles.AddSample(float64(len(of)))
	}
}

func addNetIOCounters(n *net.IOCountersStat, nioc *NetIOCounters) {
	nioc.InterfaceName = n.Name
	nioc.BytesRecv.Set(float64(n.BytesRecv))
	nioc.BytesSent.Set(float64(n.BytesSent))
	nioc.ErrIn.Set(float64(n.Errin))
	nioc.ErrOut.Set(float64(n.Errout))
	nioc.PacketsRecv.Set(float64(n.PacketsRecv))
	nioc.PacketsSent.Set(float64(n.PacketsSent))
}

func (ps *ProcessStats) Stop() {
	ps.worker.Stop()
}

//Order processeses
// type TopProcesses []*ProcessCounters

// func (t TopProcesses) Len() int      { return len(t) }
// func (t TopProcesses) Swap(i, j int) { t[i], t[j] = t[j], t[i] }

// type ProcessByCPU struct {
// 	TopProcesses
// }

// sort.Slice(people, func(i, j int) bool { return people[i].Name < people[j].Name })
// pi := CPUAvgPerc(&proc.CPUTimes.System, 10*time.Second)

func (p *ProcessStats) GetTopCPULoad() []*ProcessCounters {
	pa := p.processesArray()
	sort.Slice(pa, func(i, j int) bool {
		pi := pa[i]
		pj := pa[j]

		sysi, _ := CPUAvgPerc(&pi.CPUTimes.System, 10*time.Second)
		usri, _ := CPUAvgPerc(&pi.CPUTimes.User, 10*time.Second)
		sysj, _ := CPUAvgPerc(&pj.CPUTimes.System, 10*time.Second)
		usrj, _ := CPUAvgPerc(&pj.CPUTimes.User, 10*time.Second)

		return (sysj + usrj) < (sysi + usri)
	})
	return pa
}

func (p *ProcessStats) GetTopCPUIOWait() []*ProcessCounters {
	pa := p.processesArray()
	sort.Slice(pa, func(i, j int) bool {
		pi := pa[i]
		pj := pa[j]

		wi, _ := CPUAvgPerc(&pi.CPUTimes.IOWait, 10*time.Second)
		wj, _ := CPUAvgPerc(&pj.CPUTimes.IOWait, 10*time.Second)

		return wj < wi
	})
	return pa
}

func (p *ProcessStats) GetTopIOBytes(read bool) []*ProcessCounters {
	pa := p.processesArray()
	sort.Slice(pa, func(i, j int) bool {
		pi := pa[i]
		pj := pa[j]

		if read {
			ri, _ := CPUAvgPerc(&pi.IOCounters.ReadBytes.Timeseries, 10*time.Second)
			rj, _ := CPUAvgPerc(&pj.IOCounters.ReadBytes.Timeseries, 10*time.Second)
			return rj < ri
		}

		wi, _ := CPUAvgPerc(&pi.IOCounters.WriteBytes.Timeseries, 10*time.Second)
		wj, _ := CPUAvgPerc(&pj.IOCounters.WriteBytes.Timeseries, 10*time.Second)
		return wj < wi

	})
	return pa
}

func (p *ProcessStats) GetTopIOCount(read bool) []*ProcessCounters {
	pa := p.processesArray()
	sort.Slice(pa, func(i, j int) bool {
		pi := pa[i]
		pj := pa[j]

		if read {
			ri, _ := CPUAvgPerc(&pi.IOCounters.ReadCount.Timeseries, 10*time.Second)
			rj, _ := CPUAvgPerc(&pj.IOCounters.ReadCount.Timeseries, 10*time.Second)
			return rj < ri
		}

		wi, _ := CPUAvgPerc(&pi.IOCounters.WriteCount.Timeseries, 10*time.Second)
		wj, _ := CPUAvgPerc(&pj.IOCounters.WriteCount.Timeseries, 10*time.Second)
		return wj < wi

	})
	return pa
}

func (p *ProcessStats) GetTopMem() []*ProcessCounters {
	pa := p.processesArray()
	sort.Slice(pa, func(i, j int) bool {
		pi := pa[i]
		pj := pa[j]

		vi, _ := pi.MemoryTotal.GetLastValue()
		vj, _ := pj.MemoryTotal.GetLastValue()

		return vj.Value < vi.Value
	})
	return pa
}

func (p *ProcessStats) GetTopFD() []*ProcessCounters {
	pa := p.processesArray()
	sort.Slice(pa, func(i, j int) bool {
		pi := pa[i]
		pj := pa[j]

		vi, _ := pi.FD.GetLastValue()
		vj, _ := pj.FD.GetLastValue()

		return vj.Value < vi.Value
	})
	return pa
}

func (p *ProcessStats) GetTopNetIOBytes(recv bool) []*ProcessCounters {
	pa := p.processesArray()
	sort.Slice(pa, func(i, j int) bool {
		pi := pa[i]
		pj := pa[j]

		if recv {
			ri, _ := pi.TotalNetIOCounters.BytesRecv.Timeseries.GetLastValue()
			rj, _ := pj.TotalNetIOCounters.BytesRecv.Timeseries.GetLastValue()
			return rj.Value < ri.Value
		}

		si, _ := pi.TotalNetIOCounters.BytesSent.Timeseries.GetLastValue()
		sj, _ := pj.TotalNetIOCounters.BytesSent.Timeseries.GetLastValue()
		return sj.Value < si.Value
	})
	return pa
}

func (p *ProcessStats) GetTopNetIOErrors(in bool) []*ProcessCounters {
	pa := p.processesArray()
	sort.Slice(pa, func(i, j int) bool {
		pi := pa[i]
		pj := pa[j]

		if in {
			ri, _ := pi.TotalNetIOCounters.ErrIn.Timeseries.GetLastValue()
			rj, _ := pj.TotalNetIOCounters.ErrIn.Timeseries.GetLastValue()
			return rj.Value < ri.Value
		}

		si, _ := pi.TotalNetIOCounters.ErrOut.Timeseries.GetLastValue()
		sj, _ := pj.TotalNetIOCounters.ErrOut.Timeseries.GetLastValue()
		return sj.Value < si.Value
	})
	return pa
}

func (p *ProcessStats) processesArray() []*ProcessCounters {
	pa := make([]*ProcessCounters, 0)
	for _, v := range p.Processes {
		pa = append(pa, v)
	}
	return pa
}
