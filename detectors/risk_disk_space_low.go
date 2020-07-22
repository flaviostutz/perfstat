package detectors

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []Issue {
		issues := make([]Issue, 0)

		for pname, part := range ActiveStats.DiskStats.Partitions {

			//PARTITION USED SPACE
			total := part.Total
			free, ok := part.Free.Last()
			if !ok {
				logrus.Tracef("Not enough data for disk partition analysis")
				return []Issue{}
			}

			usedPerc := (float64(total) - free.Value) / float64(total)

			score := criticityScore(usedPerc, opt.LowDiskPercRange)
			logrus.Tracef("disk-low-space usedPerc=%.2f criticityScore=%.2f", usedPerc, score)

			if score > 0 {
				issues = append(issues, Issue{
					Typ:   "risk",
					ID:    "disk-low-space",
					Score: score,
					Res: Resource{
						Typ:           "disk",
						Name:          fmt.Sprintf("partition:%s", pname),
						PropertyName:  "space-used-perc",
						PropertyValue: usedPerc,
					},
				})
			}

			//PARTITION USED INODES
			total = part.InodesTotal
			free, ok = part.InodesFree.Last()
			if !ok {
				logrus.Tracef("Not enough data for disk partition analysis")
				return []Issue{}
			}

			usedPerc = (float64(total) - free.Value) / float64(total)

			score = criticityScore(usedPerc, opt.LowDiskPercRange)
			logrus.Tracef("disk-low-inodes usedPerc=%.2f criticityScore=%.2f", usedPerc, score)

			if score > 0 {
				issues = append(issues, Issue{
					Typ:   "risk",
					ID:    "disk-low-inodes",
					Score: score,
					Res: Resource{
						Typ:           "disk",
						Name:          fmt.Sprintf("partition:%s", pname),
						PropertyName:  "inodes-used-perc",
						PropertyValue: usedPerc,
					},
				})
			}
		}
		return issues
	})
}
