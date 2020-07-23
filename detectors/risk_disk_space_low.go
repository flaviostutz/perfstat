package detectors

import (
	"fmt"
	"time"
)

func init() {
	RegisterDetector(func(opt *Options) []DetectionResult {

		issues := make([]DetectionResult, 0)

		for pname, part := range ActiveStats.DiskStats.Partitions {

			//PARTITION USED SPACE
			r := DetectionResult{
				Typ:  "risk",
				ID:   "disk-low-space",
				When: time.Now(),
			}

			total := part.Total
			free, ok := part.Free.Last()
			if !ok {
				r.Message = notEnoughDataMessage(10 * time.Second)
				return []DetectionResult{r}
			}

			usedPerc := (float64(total) - free.Value) / float64(total)

			r.Res = Resource{
				Typ:           "disk",
				Name:          fmt.Sprintf("partition:%s", pname),
				PropertyName:  "space-used-perc",
				PropertyValue: usedPerc,
			}

			r.Score = criticityScore(usedPerc, opt.LowDiskPercRange)
			issues = append(issues, r)

			//PARTITION USED INODES
			r = DetectionResult{
				Typ:  "risk",
				ID:   "disk-low-inodes",
				When: time.Now(),
			}

			total = part.InodesTotal
			free, ok = part.InodesFree.Last()
			if !ok {
				r.Message = notEnoughDataMessage(10 * time.Second)
				return []DetectionResult{r}
			}

			usedPerc = (float64(total) - free.Value) / float64(total)

			r.Score = criticityScore(usedPerc, opt.LowDiskPercRange)
			r.Res = Resource{
				Typ:           "disk",
				Name:          fmt.Sprintf("partition:%s", pname),
				PropertyName:  "inodes-used-perc",
				PropertyValue: usedPerc,
			}

			issues = append(issues, r)
		}
		return issues
	})
}
