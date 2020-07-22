package detectors

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []Issue {
		to := time.Now()
		from := to.Add(-opt.MemLeakDuration)

		_, ok := ActiveStats.MemStats.Used.Get(from)
		if !ok {
			logrus.Debugf("Not enough timeseries data for memory leak detection. span=%v", opt.MemLeakDuration)
			return []Issue{}
		}

		_, beta, r := ActiveStats.MemStats.Used.LinearRegression(from, to)

		//linear regression error is too high
		if r < 0.4 {
			logrus.Debugf("LinearRegression for memory usage is inconclusive for memory leak detection")
			return []Issue{}
		}

		incrPerHour := beta * float64((1 * time.Hour).Nanoseconds())
		issues := make([]Issue, 0)

		//memory leak is important if growing more than 10MB per hour
		score := criticityScore(incrPerHour, [2]float64{10000000, 500000000})
		logrus.Tracef("mem-leak incrPerHour=%.2f criticityScore=%.2f", incrPerHour, score)

		if score > 0 {
			//get hungry processes with apparent mem leaks
			related := make([]Resource, 0)
			evalcount := 0
			for _, proc := range ActiveStats.ProcessStats.TopMemUsed() {
				if len(related) >= 3 || evalcount > 10 {
					break
				}

				_, ok := proc.MemoryTotal.Get(from)
				if !ok {
					logrus.Debugf("Not enough timeseries data for process memory leak detection. span=%v", opt.MemLeakDuration)
					continue
				}
				_, pbeta, r := proc.MemoryTotal.LinearRegression(from, to)
				evalcount = evalcount + 1

				//linear regression error is too high
				if r < 0.4 {
					logrus.Debugf("LinearRegression for process memory usage is inconclusive for memory leak detection")
					return []Issue{}
				}

				pincrPerHour := pbeta * float64((1 * time.Hour).Nanoseconds())
				if pincrPerHour < 10000000 {
					continue
				}

				re := Resource{
					Typ:           "process",
					Name:          fmt.Sprintf("pid:%d", proc.Pid),
					PropertyName:  "mem-hourly-bytes",
					PropertyValue: pincrPerHour,
				}
				related = append(related, re)
			}

			issues = append(issues, Issue{
				Typ:   "risk",
				ID:    "mem-leak",
				Score: score,
				Res: Resource{
					Typ:           "mem",
					Name:          "mem",
					PropertyName:  "bytes-perhour",
					PropertyValue: incrPerHour,
				},
				Related: related,
			})
		}

		return issues
	})
}
