package stats

import (
	"fmt"
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
	CPUAffinity        []int32
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
	logrus.Debugf("Process Stats: initializing...")

	ps := &ProcessStats{}

	ps.Processes = make(map[int32]*ProcessCounters)
	ps.timeseriesMaxSpan = timeseriesMaxSpan

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

	fmt.Printf(">>> %d\n", len(processes))
	for _, p := range processes {
		proc, ok := ps.Processes[p.Pid]
		if !ok {
			fmt.Printf(">>> NEW PID %d\n", p.Pid)

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
		} else {
			fmt.Printf(">>> EXISTING PROCESS\n")
		}
		ps.addProcessStats(p, proc)
	}

	return nil
}

func (ps *ProcessStats) addProcessStats(p *process.Process, proc *ProcessCounters) error {
	var err error
	proc.LastSeen = time.Now()

	//cpu affinity
	proc.CPUAffinity, err = p.CPUAffinity()
	if err != nil {
		return err
	}

	//cpu usage
	timestats, err := p.Times()
	if err != nil {
		return err
	}
	fmt.Printf("CPUTIMES = %v", proc.CPUTimes)
	proc.CPUTimes.IOWait.AddSample(timestats.Iowait)
	proc.CPUTimes.Idle.AddSample(timestats.Idle)
	proc.CPUTimes.Steal.AddSample(timestats.Steal)
	proc.CPUTimes.System.AddSample(timestats.System)
	proc.CPUTimes.User.AddSample(timestats.User)

	//network connection count
	connstats, err := p.Connections()
	if err != nil {
		return err
	}
	proc.Connections.AddSample(float64(len(connstats)))

	//network io overall
	netiostats, err := p.NetIOCounters(false)
	if err != nil {
		return err
	}
	n := &netiostats[0]
	addNetIOCounters(n, proc.TotalNetIOCounters)

	//network io per interface
	netiostats, err = p.NetIOCounters(true)
	if err != nil {
		return err
	}
	for _, niostat := range netiostats {
		nc, ok := proc.NetIOCounters[niostat.Name]
		if !ok {
			nc = &NetIOCounters{}
			nc.InterfaceName = niostat.Name
			nc.BytesRecv = signalutils.NewTimeseriesCounterRate(ps.timeseriesMaxSpan)
			nc.BytesSent = signalutils.NewTimeseriesCounterRate(ps.timeseriesMaxSpan)
			nc.ErrIn = signalutils.NewTimeseriesCounterRate(ps.timeseriesMaxSpan)
			nc.ErrOut = signalutils.NewTimeseriesCounterRate(ps.timeseriesMaxSpan)
			nc.PacketsRecv = signalutils.NewTimeseriesCounterRate(ps.timeseriesMaxSpan)
			nc.PacketsSent = signalutils.NewTimeseriesCounterRate(ps.timeseriesMaxSpan)
			proc.NetIOCounters[niostat.Name] = nc
		}
		addNetIOCounters(&niostat, nc)
	}

	//io counters
	cs, err := p.IOCounters()
	if err != nil {
		return err
	}
	ioc := proc.IOCounters
	ioc.ReadBytes.Set(float64(cs.ReadBytes))
	ioc.ReadCount.Set(float64(cs.ReadCount))
	ioc.WriteBytes.Set(float64(cs.WriteBytes))
	ioc.WriteCount.Set(float64(cs.WriteCount))

	//ram memory
	mp, err := p.MemoryPercent()
	if err != nil {
		return err
	}
	proc.MemoryPercent.AddSample(float64(mp))

	mi, err := p.MemoryInfo()
	if err != nil {
		return err
	}
	proc.MemoryTotal.AddSample(float64(mi.RSS))

	//file descriptors
	fd, err := p.NumFDs()
	if err != nil {
		return err
	}
	proc.FD.AddSample(float64(fd))

	//open files
	of, err := p.OpenFiles()
	if err != nil {
		return err
	}
	proc.OpenFiles.AddSample(float64(len(of)))

	return nil
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
