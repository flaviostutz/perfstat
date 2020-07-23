package detectors

import (
	"fmt"
	"time"

	"github.com/flaviostutz/perfstat/stats"
	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []DetectionResult {
		r := DetectionResult{
			Typ:  "bottleneck",
			ID:   "cpu-low-idle",
			When: time.Now(),
		}

		idle, ok := stats.TimeLoadPerc(&ActiveStats.CPUStats.Total.Idle, opt.CPULoadAvgDuration)
		if !ok {
			r.Message = notEnoughDataMessage(opt.CPULoadAvgDuration)
			return []DetectionResult{r}
		}

		load := 1.0 - idle

		r.Res = Resource{
			Typ:           "cpu",
			Name:          "cpu:all",
			PropertyName:  "load-perc",
			PropertyValue: load,
		}

		r.Score = criticityScore(load, opt.HighCPUPercRange)
		logrus.Tracef("cpu-low load=%.2f criticityScore=%.2f", load, r.Score)
		if r.Score == 0 {
			return []DetectionResult{r}
		}

		//get hungry processes
		r.Related = make([]Resource, 0)
		for _, proc := range ActiveStats.ProcessStats.TopCPULoad() {
			if len(r.Related) >= 3 {
				break
			}
			ps, _ := stats.TimeLoadPerc(&proc.CPUTimes.System, opt.CPULoadAvgDuration)
			pu, _ := stats.TimeLoadPerc(&proc.CPUTimes.User, opt.CPULoadAvgDuration)
			res := Resource{
				Typ:           "process",
				Name:          fmt.Sprintf("pid:%d", proc.Pid),
				PropertyName:  "cpu-load-perc",
				PropertyValue: pu + ps,
			}
			r.Related = append(r.Related, res)
		}

		//check if vm hypervisor is stealing CPU power
		steal, ok := stats.TimeLoadPerc(&ActiveStats.CPUStats.Total.Steal, opt.CPULoadAvgDuration)
		if ok && steal > 0.2 {
			res := Resource{
				Typ:           "cpu",
				Name:          "cpu:total",
				PropertyName:  "cpu-steal-perc",
				PropertyValue: steal,
			}
			r.Related = append(r.Related, res)
		}

		return []DetectionResult{r}
	})
}
