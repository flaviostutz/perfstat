package detectors

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []DetectionResult {

		r := DetectionResult{
			Typ:  "bottleneck",
			ID:   "mem-low",
			When: time.Now(),
		}

		used, ok := ActiveStats.MemStats.Used.Last()
		if !ok {
			r.Message = notEnoughDataMessage(opt.CPULoadAvgDuration)
			return []DetectionResult{r}
		}

		total := ActiveStats.MemStats.Total
		load := used.Value / float64(total)
		r.Res = Resource{
			Typ:           "mem",
			Name:          "ram",
			PropertyName:  "used-perc",
			PropertyValue: load,
		}

		r.Score = criticityScore(load, opt.HighMemPercRange)
		if r.Score == 0 {
			return []DetectionResult{r}
		}

		//get hungry processes
		r.Related = make([]Resource, 0)
		for _, proc := range ActiveStats.ProcessStats.TopMemUsed() {
			if len(r.Related) >= 5 {
				break
			}
			mt, ok := proc.MemoryTotal.Last()
			if !ok {
				logrus.Tracef("Couldn't get memory total for pid %d", proc.Pid)
				continue
			}
			res := Resource{
				Typ:           "process",
				Name:          fmt.Sprintf("pid:%d", proc.Pid),
				PropertyName:  "mem-used-bytes",
				PropertyValue: mt.Value,
			}
			r.Related = append(r.Related, res)
		}

		return []DetectionResult{r}
	})
}
