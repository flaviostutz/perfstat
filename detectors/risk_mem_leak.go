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
			ID:   "fd-low",
			When: time.Now(),
		}

		to := time.Now()
		from := to.Add(-opt.MemLeakDuration)

		_, ok := ActiveStats.MemStats.Used.Get(from)
		if !ok {
			r.Message = notEnoughDataMessage(opt.MemLeakDuration)
			r.Score = -1
			return []DetectionResult{r}
		}

		_, beta, r0 := ActiveStats.MemStats.Used.LinearRegression(from, to)
		//linear regression error is too high
		if r0 < 0.4 {
			r.Message = "Analysis is inconclusive"
			return []DetectionResult{r}
		}

		incrPerHour := beta * float64((1 * time.Hour).Nanoseconds())

		r.Res = Resource{
			Typ:           "mem",
			Name:          "mem",
			PropertyName:  "bytes-perhour",
			PropertyValue: incrPerHour,
		}

		//memory leak is important if growing more than 10MB per hour
		r.Score = criticityScore(incrPerHour, [2]float64{10000000, 500000000})

		if r.Score == 0 {
			return []DetectionResult{r}
		}

		//get hungry processes with apparent mem leaks
		r.Related = make([]Resource, 0)
		evalcount := 0
		for _, proc := range ActiveStats.ProcessStats.TopMemUsed() {
			if len(r.Related) >= 3 || evalcount > 10 {
				break
			}

			_, ok := proc.MemoryTotal.Get(from)
			if !ok {
				logrus.Debugf("Not enough timeseries data for process memory leak detection. span=%v", opt.MemLeakDuration)
				continue
			}
			_, pbeta, r0 := proc.MemoryTotal.LinearRegression(from, to)
			evalcount = evalcount + 1

			//linear regression error is too high
			if r0 < 0.4 {
				logrus.Debugf("LinearRegression for process memory usage is inconclusive for memory leak detection")
				continue
			}

			pincrPerHour := pbeta * float64((1 * time.Hour).Nanoseconds())
			if pincrPerHour < 10000000 {
				continue
			}

			re := Resource{
				Typ:           "process",
				Name:          fmt.Sprintf("%s(%s)[%d]", proc.Cmdline, proc.Name, proc.Pid),
				PropertyName:  "mem-hourly-bytes",
				PropertyValue: pincrPerHour,
			}
			r.Related = append(r.Related, re)
		}

		return []DetectionResult{r}
	})
}
