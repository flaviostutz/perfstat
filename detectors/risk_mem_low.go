package detectors

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []DetectionResult {

		r := DetectionResult{
			Typ:  "risk",
			ID:   "mem-low",
			When: time.Now(),
		}

		to := time.Now()
		from := to.Add(-opt.MemAvgDuration)

		mtotal := ActiveStats.MemStats.Total
		stotal := ActiveStats.MemStats.SwapTotal

		mused, ok := ActiveStats.MemStats.Used.Avg(time.Now().Add(-opt.MemAvgDuration), time.Now())
		if !ok {
			r.Message = notEnoughDataMessage(opt.MemLeakDuration)
			r.Score = -1
			return []DetectionResult{r}
		}
		sused, ok := ActiveStats.MemStats.SwapUsed.Avg(time.Now().Add(-opt.MemAvgDuration), time.Now())
		if !ok {
			r.Message = notEnoughDataMessage(opt.MemLeakDuration)
			r.Score = -1
			return []DetectionResult{r}
		}

		usedPerc := float64(mused+sused) / float64(mtotal+stotal)

		r.Res = Resource{
			Typ:           "mem",
			Name:          "mem",
			PropertyName:  "used-perc",
			PropertyValue: usedPerc,
		}

		r.Score = criticityScore(usedPerc, opt.HighMemPercRange)

		if r.Score == 0 {
			return []DetectionResult{r}
		}

		//get hungry processes
		r.Related = make([]Resource, 0)
		for _, proc := range ActiveStats.ProcessStats.TopMemUsed() {
			if len(r.Related) >= 3 {
				break
			}
			pused, ok := proc.MemoryTotal.Avg(from, to)
			if !ok {
				logrus.Tracef("Couldn't get used mem for pid %d", proc.Pid)
				continue
			}
			res := Resource{
				Typ:           "process",
				Name:          fmt.Sprintf("pid:%d", proc.Pid),
				PropertyName:  "mem-used-bytes",
				PropertyValue: pused,
			}
			r.Related = append(r.Related, res)
		}

		//verify too much swap usage
		sUsedPerc := sused / float64(stotal)
		if sUsedPerc > 0.7 {
			r1 := Resource{
				Typ:           "mem",
				Name:          "swap",
				PropertyName:  "mem-swap-perc",
				PropertyValue: sUsedPerc,
			}
			r.Related = append(r.Related, r1)
		}

		//verify too much ram usage
		rUsedPerc := sused / float64(stotal)
		if sUsedPerc > 0.7 {
			r1 := Resource{
				Typ:           "mem",
				Name:          "ram",
				PropertyName:  "mem-ram-perc",
				PropertyValue: rUsedPerc,
			}
			r.Related = append(r.Related, r1)
		}

		return []DetectionResult{r}
	})
}
