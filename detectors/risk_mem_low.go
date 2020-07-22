package detectors

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []Issue {
		to := time.Now()
		from := to.Add(-opt.MemAvgDuration)

		mtotal := ActiveStats.MemStats.Total
		stotal := ActiveStats.MemStats.SwapTotal

		mused, ok := ActiveStats.MemStats.Used.Avg(time.Now().Add(-opt.MemAvgDuration), time.Now())
		if !ok {
			logrus.Tracef("Not enough data for mem analysis")
			return []Issue{}
		}
		sused, ok := ActiveStats.MemStats.SwapUsed.Avg(time.Now().Add(-opt.MemAvgDuration), time.Now())
		if !ok {
			logrus.Tracef("Not enough data for mem analysis")
			return []Issue{}
		}

		usedPerc := float64(mused+sused) / float64(mtotal+stotal)

		issues := make([]Issue, 0)

		score := criticityScore(usedPerc, opt.HighMemPercRange)
		logrus.Tracef("mem-low totalUsedPerc=%.2f criticityScore=%.2f", usedPerc, score)

		if score > 0 {
			related := make([]Resource, 0)

			//get hungry processes
			for _, proc := range ActiveStats.ProcessStats.TopMemUsed() {
				if len(related) >= 3 {
					break
				}
				pused, ok := proc.MemoryTotal.Avg(from, to)
				if !ok {
					logrus.Tracef("Couldn't get used mem for pid %d", proc.Pid)
					continue
				}
				r := Resource{
					Typ:           "process",
					Name:          fmt.Sprintf("pid:%d", proc.Pid),
					PropertyName:  "mem-used-bytes",
					PropertyValue: pused,
				}
				related = append(related, r)
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
				related = append(related, r1)
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
				related = append(related, r1)
			}

			issues = append(issues, Issue{
				Typ:   "risk",
				ID:    "mem-low",
				Score: score,
				Res: Resource{
					Typ:           "mem",
					Name:          "mem",
					PropertyName:  "used-perc",
					PropertyValue: usedPerc,
				},
				Related: related,
			})
		}

		return issues
	})
}
