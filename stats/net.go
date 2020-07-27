package stats

import (
	"context"
	"sort"
	"time"

	"github.com/flaviostutz/signalutils"
	"github.com/shirou/gopsutil/net"
	"github.com/sirupsen/logrus"
)

type NetStats struct {
	NICs               map[string]*NICMetrics
	timeseriesSize     time.Duration
	ioRateLoadDuration time.Duration
}

type NICMetrics struct {
	Name        string
	BytesRecv   signalutils.TimeseriesCounterRate
	BytesSent   signalutils.TimeseriesCounterRate
	PacketsRecv signalutils.TimeseriesCounterRate
	PacketsSent signalutils.TimeseriesCounterRate
	ErrIn       signalutils.TimeseriesCounterRate
	ErrOut      signalutils.TimeseriesCounterRate
}

func NewNetStats(ctx context.Context, timeseriesSize time.Duration, ioRateLoadDuration time.Duration, sampleFreq float64) *NetStats {
	logrus.Tracef("Net Stats: initializing...")

	d := &NetStats{
		NICs:               make(map[string]*NICMetrics),
		timeseriesSize:     timeseriesSize,
		ioRateLoadDuration: ioRateLoadDuration,
	}

	signalutils.StartWorker(ctx, "net", d.netStep, sampleFreq/2, sampleFreq, true)
	logrus.Debugf("Net Stats: running")

	return d
}

func (d *NetStats) netStep() error {

	//stats per nic
	ioc, err := net.IOCounters(true)
	if err != nil {
		return err
	}

	for _, is := range ioc {
		nm, ok := d.NICs[is.Name]

		if !ok {
			nm = &NICMetrics{
				Name:        is.Name,
				BytesRecv:   signalutils.NewTimeseriesCounterRate(d.timeseriesSize),
				BytesSent:   signalutils.NewTimeseriesCounterRate(d.timeseriesSize),
				PacketsRecv: signalutils.NewTimeseriesCounterRate(d.timeseriesSize),
				PacketsSent: signalutils.NewTimeseriesCounterRate(d.timeseriesSize),
				ErrIn:       signalutils.NewTimeseriesCounterRate(d.timeseriesSize),
				ErrOut:      signalutils.NewTimeseriesCounterRate(d.timeseriesSize),
			}
			d.NICs[is.Name] = nm
		}

		//add stats to timeseries
		nm.BytesRecv.Set(float64(is.BytesRecv))
		nm.BytesSent.Set(float64(is.BytesSent))
		nm.PacketsRecv.Set(float64(is.PacketsRecv))
		nm.PacketsSent.Set(float64(is.PacketsSent))
		nm.ErrIn.Set(float64(is.Dropin + is.Errin))
		nm.ErrOut.Set(float64(is.Dropout + is.Errout))
	}

	return nil
}

func (d *NetStats) TopPacketRate(recv bool) []*NICMetrics {
	da := d.nicArray()
	sort.Slice(da, func(i, j int) bool {
		pi := da[i]
		pj := da[j]

		if recv {
			ri, _ := pi.PacketsRecv.Rate(d.ioRateLoadDuration)
			rj, _ := pj.PacketsRecv.Rate(d.ioRateLoadDuration)
			return rj < ri
		}

		wi, _ := pi.PacketsSent.Rate(d.ioRateLoadDuration)
		wj, _ := pj.PacketsSent.Rate(d.ioRateLoadDuration)
		return wj < wi
	})
	return da
}

func (d *NetStats) TopByteRate(recv bool) []*NICMetrics {
	da := d.nicArray()
	sort.Slice(da, func(i, j int) bool {
		pi := da[i]
		pj := da[j]

		if recv {
			ri, _ := pi.BytesRecv.Rate(d.ioRateLoadDuration)
			rj, _ := pj.BytesRecv.Rate(d.ioRateLoadDuration)
			return rj < ri
		}

		wi, _ := pi.BytesSent.Rate(d.ioRateLoadDuration)
		wj, _ := pj.BytesSent.Rate(d.ioRateLoadDuration)
		return wj < wi
	})
	return da
}

func (d *NetStats) TopErrorsRate(in bool) []*NICMetrics {
	da := d.nicArray()
	sort.Slice(da, func(i, j int) bool {
		pi := da[i]
		pj := da[j]

		if in {
			ri, _ := pi.ErrIn.Rate(d.ioRateLoadDuration)
			rj, _ := pj.ErrIn.Rate(d.ioRateLoadDuration)
			return rj < ri
		}

		wi, _ := pi.ErrOut.Rate(d.ioRateLoadDuration)
		wj, _ := pj.ErrOut.Rate(d.ioRateLoadDuration)
		return wj < wi
	})
	return da
}

func (d *NetStats) nicArray() []*NICMetrics {
	dms := make([]*NICMetrics, 0)
	for _, v := range d.NICs {
		dms = append(dms, v)
	}
	return dms
}
