package perfstat

import (
	"fmt"
	"sort"
	"time"

	"github.com/flaviostutz/perfstat/detectors"
	"github.com/flaviostutz/perfstat/stats"
	"github.com/flaviostutz/signalutils"
	"github.com/sirupsen/logrus"
	"github.com/yaacov/observer/observer"
)

//Perfstat performance analyser
type Perfstat struct {
	opt          detectors.Options
	cpuStats     *stats.CPUStats
	processStats *stats.ProcessStats
	worker1      *signalutils.Worker
	worker2      *signalutils.Worker
	curResults   []detectors.DetectionResult
	observer     observer.Observer
}

type IssueEvent struct {
	When  time.Time
	Issue detectors.DetectionResult
	Typ   string
}

//Start initializes a new Perfstat utility
func Start(opt detectors.Options) *Perfstat {
	p := &Perfstat{
		opt: opt,
	}

	logrus.Debugf("Starting detectors")
	detectors.Start(opt)
	time.Sleep(1 * time.Second)

	logrus.Debugf("Starting issues tracker")
	p.worker1 = signalutils.StartWorker("perfstat-detect", func() error {
		result, err := p.DetectNow()
		if err != nil {
			return err
		}

		p.observer = observer.Observer{}
		p.observer.Open()
		p.observer.Emit(result)

		p.curResults = result
		return nil
	}, 0.5, 1, true)

	return p
}

func (p *Perfstat) SetLogLevel(level logrus.Level) {
	logrus.SetLevel(level)
	detectors.SetLogLevel(level)
}

func (p *Perfstat) Watch(issueEvents chan IssueEvent) {
	p.observer.AddListener(func(e interface{}) {
		iss := e.(IssueEvent)
		issueEvents <- iss
	})
}

func (p *Perfstat) TopCriticity() []detectors.DetectionResult {
	da := make([]detectors.DetectionResult, 0)
	for _, v := range p.curResults {
		da = append(da, v)
	}
	sort.Slice(da, func(i, j int) bool {
		si := da[i].Score
		sj := da[j].Score
		if si == sj {
			return da[i].Res.PropertyValue > da[j].Res.PropertyValue
		}
		return si > sj
	})
	return da
}

//DetectNow perform issues detection on the system once
func (p *Perfstat) DetectNow() ([]detectors.DetectionResult, error) {
	results := make([]detectors.DetectionResult, 0)
	if !detectors.Started {
		return []detectors.DetectionResult{}, fmt.Errorf("Perfstat not started yet")
	}
	logrus.Debugf("Perfstat DetectNow()")
	for _, df := range detectors.DetectorFuncs {
		r := df(&detectors.Opt)
		// for _, iss := range r {
		// 	logrus.Debugf("RESULT: %s", iss.String())
		// }
		results = append(results, r...)
	}
	return results, nil
}

func (p *Perfstat) Stop() {
	detectors.Stop()
}
