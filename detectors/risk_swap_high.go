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

		sin, ok := ActiveStats.MemStats.SwapIn.Rate(opt.MemAvgDuration)
		if !ok {
			logrus.Debugf("Not enough timeseries data for swap analysis. span=%v", opt.MemLeakDuration)
			return []Issue{}
		}
		sout, ok := ActiveStats.MemStats.SwapOut.Rate(opt.MemAvgDuration)
		if !ok {
			logrus.Debugf("Not enough timeseries data for swap analysis. span=%v", opt.MemLeakDuration)
			return []Issue{}
		}

		issues := make([]Issue, 0)

		score := criticityScore(sin+sout, opt.HighSwapBpsRange)
		logrus.Tracef("mem-swap-bps score=%.2f criticityScore=%.2f", sin+sout, score)

		if score > 0 {
			//get processes with high swap
			related := make([]Resource, 0)
			for _, proc := range ActiveStats.ProcessStats.TopMemSwap() {
				if len(related) >= 5 {
					break
				}

				swap, ok := proc.MemorySwap.Avg(from, to)
				if !ok {
					continue
				}
				re := Resource{
					Typ:           "process",
					Name:          fmt.Sprintf("pid:%d", proc.Pid),
					PropertyName:  "swap-bytes",
					PropertyValue: swap,
				}
				related = append(related, re)
			}

			issues = append(issues, Issue{
				Typ:   "risk",
				ID:    "mem-swap-high",
				Score: score,
				Res: Resource{
					Typ:           "mem",
					Name:          "swap",
					PropertyName:  "swap-total-bps",
					PropertyValue: sin + sout,
				},
				Related: related,
			})
		}

		return issues
	})
}
