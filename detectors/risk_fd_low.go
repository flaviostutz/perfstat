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
		from := to.Add(-opt.IORateLoadDuration)

		usedFD, ok := ActiveStats.DiskStats.FD.UsedFD.Avg(from, to)
		if !ok {
			r.Message = notEnoughDataMessage(opt.IORateLoadDuration)
			r.Score = -1
			return []DetectionResult{r}
		}
		maxFD := ActiveStats.DiskStats.FD.MaxFD
		fdUsedPerc := float64(usedFD) / float64(maxFD)

		r.Score = criticityScore(fdUsedPerc, opt.FDUsedRange)
		r.Res = Resource{
			Typ:           "fd",
			Name:          "fd",
			PropertyName:  "used-perc",
			PropertyValue: fdUsedPerc,
		}

		if r.Score == 0 {
			return []DetectionResult{r}
		}

		//get hungry processes
		r.Related = make([]Resource, 0)
		for _, proc := range ActiveStats.ProcessStats.TopFD() {
			if len(r.Related) >= 3 {
				break
			}
			pused, ok := proc.FD.Avg(from, to)
			if !ok {
				logrus.Tracef("Couldn't get used fd for pid %d", proc.Pid)
				continue
			}
			res := Resource{
				Typ:           "process",
				Name:          fmt.Sprintf("%s(%s)[%d]", proc.Cmdline, proc.Name, proc.Pid),
				PropertyName:  "fd-used-count",
				PropertyValue: pused,
			}
			r.Related = append(r.Related, res)
		}

		return []DetectionResult{r}
	})
}
