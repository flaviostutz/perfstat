package perfstat

import (
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/coryb/sorty"
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

//TopCriticity returns the most important items found in system
//if typ or id is "", all results are returned
//removeNear is used to hide occurrences that have similar contents in id, score and prop value
func (p *Perfstat) TopCriticity(minScore float64, typ string, idRegex string, removeNear bool) []detectors.DetectionResult {

	data := make([]map[string]interface{}, 0)
	for _, v := range p.curResults {
		d := map[string]interface{}{
			"ref": v,
			// "Typ":                  v.Typ,
			// "ID":                   v.ID,
			"Score":    round(v.Score),
			"Res.Name": v.Res.Name,
			// "Res.Typ":              v.Res.Typ,
			"Res.PropertyName":  v.Res.PropertyName,
			"Res.PropertyValue": round(v.Res.PropertyValue),
		}
		data = append(data, d)
	}

	s := sorty.NewSorter().ByKeys([]string{
		"-Score",
		"-Res.PropertyValue",
		"+Res.Name",
		"+Res.PropertyName",
	})
	s.Sort(data)

	da := make([]detectors.DetectionResult, 0)
	for _, v := range data {
		o := v["ref"].(detectors.DetectionResult)
		if typ != "" && o.Typ != typ {
			continue
		}
		if idRegex != "" {
			re := regexp.MustCompile(idRegex)
			if !re.MatchString(o.ID) {
				continue
			}
		}
		if o.Score < minScore {
			continue
		}
		da = append(da, o)
	}

	if removeNear {
		rr := make([]detectors.DetectionResult, 0)
		p := detectors.DetectionResult{}
		for _, v := range da {
			add := true
			if p.Typ == v.Typ && p.ID == v.ID {
				if near(p.Score, v.Score, 0.01) {
					if near(p.Res.PropertyValue, v.Res.PropertyValue, 0.01) {
						add = false
					}
				}
			}
			if add {
				rr = append(rr, v)
			}
			p = v
		}
		return rr
	}

	return da
}

func (p *Perfstat) Score(typ string, idRegex string) float64 {
	dr := p.TopCriticity(0.0, typ, idRegex, false)
	v := 0.0
	for _, r := range dr {
		if r.Score > v {
			v = r.Score
		}
	}
	return v
}

//DetectNow perform issues detection on the system once
func (p *Perfstat) DetectNow() ([]detectors.DetectionResult, error) {
	results := make([]detectors.DetectionResult, 0)
	if !detectors.Started {
		return []detectors.DetectionResult{}, fmt.Errorf("Perfstat not started yet")
	}
	// logrus.Debugf("Perfstat DetectNow()")
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

func round(x float64) float64 {
	return math.Round(x*100) / 100
}
func near(x1 float64, x2 float64, distPerc float64) bool {
	if x1 == 0 {
		return true
	}
	return (math.Abs(x1-x2) / x1) < distPerc
}
