package stats

import (
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/flaviostutz/signalutils"
	"github.com/shirou/gopsutil/disk"
	"github.com/sirupsen/logrus"
)

type DiskStats struct {
	Disks              map[string]*DiskMetrics
	Partitions         map[string]*PartitionMetrics
	FD                 *FDMetrics
	worker             *signalutils.Worker
	timeseriesSize     time.Duration
	ioRateLoadDuration time.Duration
}

type FDMetrics struct {
	UsedFD signalutils.Timeseries
	MaxFD  int64
}

type DiskMetrics struct {
	Name           string
	SerialNumber   string
	IoTime         signalutils.Timeseries
	ReadBytes      signalutils.TimeseriesCounterRate
	ReadCount      signalutils.TimeseriesCounterRate
	ReadTime       signalutils.Timeseries
	WriteBytes     signalutils.TimeseriesCounterRate
	WriteCount     signalutils.TimeseriesCounterRate
	WriteTime      signalutils.Timeseries
	IopsInProgress signalutils.Timeseries
}

type PartitionMetrics struct {
	Path        string
	Fstype      string
	Total       uint64
	Free        signalutils.Timeseries
	InodesTotal uint64
	InodesFree  signalutils.Timeseries
}

func NewDiskStats(timeseriesSize time.Duration, ioRateLoadDuration time.Duration, sampleFreq float64) *DiskStats {
	logrus.Tracef("Disk Stats: initializing...")

	d := &DiskStats{
		Disks:      make(map[string]*DiskMetrics),
		Partitions: make(map[string]*PartitionMetrics),
		FD: &FDMetrics{
			UsedFD: signalutils.NewTimeseries(timeseriesSize),
			MaxFD:  52427,
		},
		timeseriesSize:     timeseriesSize,
		ioRateLoadDuration: ioRateLoadDuration,
	}

	d.worker = signalutils.StartWorker("disk", d.diskStep, sampleFreq/2, sampleFreq, true)
	logrus.Debugf("Disk Stats: running")

	return d
}

func (d *DiskStats) diskStep() error {

	//fd stats
	usedFD, maxFD, err := FDStats()
	if err != nil {
		logrus.Tracef("FD stats works only on Linux systems. err=%s", err)
	} else {
		d.FD.MaxFD = maxFD
		d.FD.UsedFD.Add(float64(usedFD))
	}

	//stats per disk
	ioc, err := disk.IOCounters()
	if err != nil {
		return err
	}

	for name, is := range ioc {
		dm, ok := d.Disks[name]

		if !ok {
			dm = &DiskMetrics{
				Name:           name,
				SerialNumber:   is.SerialNumber,
				IoTime:         signalutils.NewTimeseries(d.timeseriesSize),
				IopsInProgress: signalutils.NewTimeseries(d.timeseriesSize),
				ReadBytes:      signalutils.NewTimeseriesCounterRate(d.timeseriesSize),
				ReadCount:      signalutils.NewTimeseriesCounterRate(d.timeseriesSize),
				ReadTime:       signalutils.NewTimeseries(d.timeseriesSize),
				WriteBytes:     signalutils.NewTimeseriesCounterRate(d.timeseriesSize),
				WriteCount:     signalutils.NewTimeseriesCounterRate(d.timeseriesSize),
				WriteTime:      signalutils.NewTimeseries(d.timeseriesSize),
			}

			d.Disks[name] = dm
		}

		//add stats to timeseries
		dm.IoTime.Add(float64(is.IoTime))
		dm.IopsInProgress.Add(float64(is.IopsInProgress))
		dm.ReadBytes.Set(float64(is.ReadBytes))
		dm.ReadCount.Set(float64(is.ReadCount))
		dm.ReadTime.Add(float64(is.ReadTime))
		dm.WriteBytes.Set(float64(is.WriteBytes))
		dm.WriteCount.Set(float64(is.WriteCount))
		dm.WriteTime.Add(float64(is.WriteTime))
	}

	//stats per partition
	partitions, err := disk.Partitions(true)
	if err != nil {
		return err
	}

	for _, p := range partitions {
		pm, ok := d.Partitions[p.Mountpoint]

		if !ok {
			pm = &PartitionMetrics{
				Path:        p.Mountpoint,
				Fstype:      p.Fstype,
				Total:       0,
				Free:        signalutils.NewTimeseries(d.timeseriesSize),
				InodesTotal: 0,
				InodesFree:  signalutils.NewTimeseries(d.timeseriesSize),
			}
			d.Partitions[p.Mountpoint] = pm
		}

		//add stats to timeseries
		pu, err := disk.Usage(p.Mountpoint)
		if err != nil {
			return err
		}
		pm.Free.Add(float64(pu.Free))
		pm.Total = pu.Total
		pm.InodesFree.Add(float64(pu.InodesFree))
		pm.InodesTotal = pu.InodesTotal
	}

	return nil
}

func (d *DiskStats) TopOpRate(read bool) []*DiskMetrics {
	da := d.diskArray()
	sort.Slice(da, func(i, j int) bool {
		pi := da[i]
		pj := da[j]

		if read {
			ri, _ := pi.ReadCount.Rate(d.ioRateLoadDuration)
			rj, _ := pj.ReadCount.Rate(d.ioRateLoadDuration)
			return rj < ri
		}

		wi, _ := pi.WriteCount.Rate(d.ioRateLoadDuration)
		wj, _ := pj.WriteCount.Rate(d.ioRateLoadDuration)
		return wj < wi
	})
	return da
}

func (d *DiskStats) TopByteRate(read bool) []*DiskMetrics {
	da := d.diskArray()
	sort.Slice(da, func(i, j int) bool {
		pi := da[i]
		pj := da[j]

		if read {
			ri, _ := pi.ReadBytes.Rate(d.ioRateLoadDuration)
			rj, _ := pj.ReadBytes.Rate(d.ioRateLoadDuration)
			return rj < ri
		}

		wi, _ := pi.WriteBytes.Rate(d.ioRateLoadDuration)
		wj, _ := pj.WriteBytes.Rate(d.ioRateLoadDuration)
		return wj < wi
	})
	return da
}

func (d *DiskStats) TopIOUtil(read bool) []*DiskMetrics {
	da := d.diskArray()
	sort.Slice(da, func(i, j int) bool {
		pi := da[i]
		pj := da[j]

		if read {
			ri, _ := TimeLoadPerc(&pi.ReadTime, d.ioRateLoadDuration)
			rj, _ := TimeLoadPerc(&pj.ReadTime, d.ioRateLoadDuration)
			return rj < ri
		}

		wi, _ := TimeLoadPerc(&pi.WriteTime, d.ioRateLoadDuration)
		wj, _ := TimeLoadPerc(&pj.WriteTime, d.ioRateLoadDuration)
		return wj < wi
	})
	return da
}

func (d *DiskStats) diskArray() []*DiskMetrics {
	dms := make([]*DiskMetrics, 0)
	for _, v := range d.Disks {
		dms = append(dms, v)
	}
	return dms
}

func (d *DiskStats) partitionArray() []*PartitionMetrics {
	dms := make([]*PartitionMetrics, 0)
	for _, v := range d.Partitions {
		dms = append(dms, v)
	}
	return dms
}

func FDStats() (usedFD int64, maxFD int64, err error) {
	filenrb, err := ioutil.ReadFile("/proc/sys/fs/file-nr")
	if err != nil {
		return -1, -1, err
	}
	fm := string(filenrb)
	filenr := strings.Fields(fm)

	fr0, err := strconv.ParseInt(filenr[0], 10, 64)
	if err != nil {
		return -1, -1, err
	}
	fr1, err := strconv.ParseInt(filenr[1], 10, 64)
	if err != nil {
		return -1, -1, err
	}
	fr2, err := strconv.ParseInt(filenr[2], 10, 64)
	if err != nil {
		return -1, -1, err
	}
	return (fr0 - fr1), fr2, nil
}

func (d *DiskStats) Stop() {
	d.worker.Stop()
}
