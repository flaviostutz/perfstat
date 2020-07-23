package detectors

import (
	"fmt"
	"math"
	"time"

	"github.com/flaviostutz/perfstat/stats"
	"github.com/flaviostutz/signalutils"
	"github.com/sirupsen/logrus"
)

var (
	Started     = false
	Opt         Options
	ActiveStats *StatsType
)

type StatsType struct {
	CPUStats     *stats.CPUStats
	ProcessStats *stats.ProcessStats
	MemStats     *stats.MemStats
	DiskStats    *stats.DiskStats
	NetStats     *stats.NetStats
}

//NewOptions create a new default options
func NewOptions() Options {
	return Options{
		HighCPUPercRange:        [2]float64{0.70, 0.95},
		HighCPUWaitPercRange:    [2]float64{0.05, 0.50},
		HighMemPercRange:        [2]float64{0.70, 0.95},
		HighSwapBpsRange:        [2]float64{10000000, 100000000},
		LowDiskPercRange:        [2]float64{0.70, 0.90},
		HighDiskUtilPercRange:   [2]float64{0.70, 0.90},
		LowFileHandlesPercRange: [2]float64{0.70, 0.90},
		FDUsedRange:             [2]float64{0.6, 0.9},
		NICErrorsRange:          [2]float64{1, 10},
		DefaultSampleFreq:       1.0,
		DefaultTimeseriesSize:   30 * time.Minute,
		CPULoadAvgDuration:      1 * time.Minute,
		IORateLoadDuration:      1 * time.Minute,
		IOLimitsSpan:            1 * time.Minute,
		MemAvgDuration:          1 * time.Minute,
		MemLeakDuration:         20 * time.Minute,
	}
}

//Options performance analysis options
type Options struct {
	Loglevel                string
	HighCPUPercRange        [2]float64
	HighCPUWaitPercRange    [2]float64
	HighMemPercRange        [2]float64
	LowDiskPercRange        [2]float64
	LowFileHandlesPercRange [2]float64
	FDUsedRange             [2]float64
	NICErrorsRange          [2]float64
	HighSwapBpsRange        [2]float64
	HighDiskUtilPercRange   [2]float64
	DefaultSampleFreq       float64
	DefaultTimeseriesSize   time.Duration
	CPULoadAvgDuration      time.Duration
	IORateLoadDuration      time.Duration
	MemAvgDuration          time.Duration
	MemLeakDuration         time.Duration
	IOLimitsSpan            time.Duration
}

//Resource a computational resource
type Resource struct {
	//Typ resource type. ex: ram, disk, cpu, fs, net
	Typ string
	//Name identification of the resource in the system. ex: /dev/fda021, cpu:23, process:prometheus[6456]
	Name string
	//PropertyName name of the resource property. ex: available-cpu, disk-writes-per-second
	PropertyName string
	//PropertyValue value of the resource property
	PropertyValue float64
}

func (r *Resource) String() string {
	return fmt.Sprintf("type=%s name=%s prop=%s propv=%.2f", r.Typ, r.Name, r.PropertyName, r.PropertyValue)
}

//DetectionResult detection results
type DetectionResult struct {
	//Typ type of issue: bottleneck, risk, harm, lib-error
	Typ string
	//Id name of the issue. ex: low-available-cpu, disk-write-ceil
	ID string
	//Score how critical is this issue to the health of the system
	Score float64
	//Message text indicating the issue details
	Message string
	//Res resource directly related to the issue
	Res Resource
	//Related related resources to the issue (ex: for low CPU, place top 3 CPU processes)
	Related []Resource
	//InfoURL contains more information on how to deal with the issue
	InfoURL string
	//When time of detection process
	When time.Time
}

func (i *DetectionResult) String() string {
	return fmt.Sprintf("type=%s id=%s score=%.2f resource=[%s] message=%s infoURL=%s", i.Typ, i.ID, i.Score, i.Res.String(), i.Message, i.InfoURL)
}

//DetectorFunc function that is called for detecting issues on the system
type DetectorFunc func(*Options) []DetectionResult

//DetectorFuncs detector functions
var DetectorFuncs = make([]DetectorFunc, 0)

//RegisterDetector register a new function to be called for detecting issues on the system
func RegisterDetector(d DetectorFunc) {
	logrus.Debugf("Registering detector %v", d)
	DetectorFuncs = append(DetectorFuncs, d)
}

//calculates a score between 0-1. 0 is "no worry"; 1 is "IT BROKE!"
func criticityScore(value float64, criticityRange [2]float64) float64 {
	if value < criticityRange[0] {
		return 0
	}
	return math.Min((value-criticityRange[0])/(criticityRange[1]-criticityRange[0]), 1.0)
}

func SetLogLevel(level logrus.Level) {
	logrus.SetLevel(level)
}

func Start(opt Options) {
	if !Started {
		ActiveStats = &StatsType{}
		ActiveStats.CPUStats = stats.NewCPUStats(opt.DefaultTimeseriesSize, 1)
		ActiveStats.ProcessStats = stats.NewProcessStats(opt.DefaultTimeseriesSize, opt.IORateLoadDuration, opt.CPULoadAvgDuration, opt.MemAvgDuration, opt.DefaultSampleFreq)
		ActiveStats.MemStats = stats.NewMemStats(opt.DefaultTimeseriesSize, opt.DefaultSampleFreq)
		ActiveStats.DiskStats = stats.NewDiskStats(opt.DefaultTimeseriesSize, opt.IORateLoadDuration, opt.DefaultSampleFreq)
		ActiveStats.NetStats = stats.NewNetStats(opt.DefaultTimeseriesSize, opt.IORateLoadDuration, opt.DefaultSampleFreq)
		Opt = opt
		Started = true
	}
}

func Stop() {
	if Started {
		ActiveStats.CPUStats.Stop()
		ActiveStats.ProcessStats.Stop()
		ActiveStats.MemStats.Stop()
		ActiveStats.DiskStats.Stop()
		ActiveStats.NetStats.Stop()
		Started = false
	}
}

func upperRateBoundaries(tcr *signalutils.TimeseriesCounterRate, from time.Time, to time.Time, opt *Options, meanHigher float64, criticityRange [2]float64) (cscore float64, meanRate float64) {
	ts, ok := tcr.RateOverTime(opt.IORateLoadDuration, opt.IOLimitsSpan)
	if !ok {
		return 0, 0
	}
	sd, m := ts.StdDev(from, to)
	if m > meanHigher {
		stdDev := (sd / m)
		return criticityScore(1.0-stdDev, criticityRange), m
	}
	return 0.0, m
}

func notEnoughDataMessage(duration time.Duration) string {
	return fmt.Sprintf("Not enough data for evaluation. min=%s", duration.String())
}
