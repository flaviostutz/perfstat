package detectors

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []Issue {
		to := time.Now()
		fromLimit := to.Add(-opt.IOLimitsSpan)

		issues := make([]Issue, 0)

		for dname, dm := range ActiveStats.DiskStats.Disks {

			//TODO add Dropped packets as a catalyser for this analysis?

			//DISK LIMIT ON WRITE BPS
			score, mean := upperRateBoundaries(&dm.WriteBytes, fromLimit, to, opt, 100000.0, [2]float64{0.8, 0.9})
			if score > 0 {
				logrus.Tracef("disk-limit-wbps bps=%.2f criticityScore=%.2f", mean, score)

				//get hungry processes
				related := make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopIOByteRate(false) {
					if len(related) >= 3 {
						break
					}
					rate, ok := proc.IOCounters.WriteBytes.Rate(opt.IORateLoadDuration)
					if !ok {
						logrus.Warnf("Couldn't get process write bps for disk on limits analysis")
						continue
					}
					if rate < 10000 {
						break
					}
					r := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("pid:%d", proc.Pid),
						PropertyName:  "disk-write-bps",
						PropertyValue: rate,
					}
					related = append(related, r)
				}

				issues = append(issues, Issue{
					Typ:   "bottleneck",
					ID:    "disk-limit-wbps",
					Score: score,
					Res: Resource{
						Typ:           "disk",
						Name:          fmt.Sprintf("disk:%s", dname),
						PropertyName:  "write-bps",
						PropertyValue: mean,
					},
					Related: related,
				})
			}

			//DISK LIMIT ON READ BPS
			score, mean = upperRateBoundaries(&dm.ReadBytes, fromLimit, to, opt, 100000.0, [2]float64{0.8, 0.9})
			if score > 0 {
				logrus.Tracef("disk-limit-rbps bps=%.2f criticityScore=%.2f", mean, score)

				//get hungry processes
				related := make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopIOByteRate(true) {
					if len(related) >= 3 {
						break
					}
					rate, ok := proc.IOCounters.ReadBytes.Rate(opt.IORateLoadDuration)
					if !ok {
						logrus.Warnf("Couldn't get process read bps for disk on limits analysis")
						continue
					}
					if rate < 10000 {
						break
					}
					r := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("pid:%d", proc.Pid),
						PropertyName:  "disk-read-bps",
						PropertyValue: rate,
					}
					related = append(related, r)
				}

				issues = append(issues, Issue{
					Typ:   "bottleneck",
					ID:    "disk-limit-rbps",
					Score: score,
					Res: Resource{
						Typ:           "disk",
						Name:          fmt.Sprintf("disk:%s", dname),
						PropertyName:  "read-bps",
						PropertyValue: mean,
					},
					Related: related,
				})
			}

			//DISK LIMIT ON WRITE OPS
			score, mean = upperRateBoundaries(&dm.WriteCount, fromLimit, to, opt, 100000.0, [2]float64{0.8, 0.9})
			if score > 0 {
				logrus.Tracef("disk-limit-wops ops=%.2f criticityScore=%.2f", mean, score)

				//get hungry processes
				related := make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopIOOpRate(false) {
					if len(related) >= 3 {
						break
					}
					rate, ok := proc.IOCounters.WriteCount.Rate(opt.IORateLoadDuration)
					if !ok {
						logrus.Warnf("Couldn't get process write ops for disk on limits analysis")
						continue
					}
					if rate < 10 {
						break
					}
					r := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("pid:%d", proc.Pid),
						PropertyName:  "disk-write-ops",
						PropertyValue: rate,
					}
					related = append(related, r)
				}

				issues = append(issues, Issue{
					Typ:   "bottleneck",
					ID:    "disk-limit-wops",
					Score: score,
					Res: Resource{
						Typ:           "disk",
						Name:          fmt.Sprintf("disk:%s", dname),
						PropertyName:  "write-ops",
						PropertyValue: mean,
					},
				})
			}

			//DISK LIMIT ON READ OPS
			score, mean = upperRateBoundaries(&dm.ReadCount, fromLimit, to, opt, 100000.0, [2]float64{0.8, 0.9})
			if score > 0 {
				logrus.Tracef("disk-limit-rops ops=%.2f criticityScore=%.2f", mean, score)

				//get hungry processes
				related := make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopIOOpRate(true) {
					if len(related) >= 3 {
						break
					}
					rate, ok := proc.IOCounters.WriteCount.Rate(opt.IORateLoadDuration)
					if !ok {
						logrus.Warnf("Couldn't get process read ops for disk on limits analysis")
						continue
					}
					if rate < 10 {
						break
					}
					r := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("pid:%d", proc.Pid),
						PropertyName:  "disk-read-ops",
						PropertyValue: rate,
					}
					related = append(related, r)
				}

				issues = append(issues, Issue{
					Typ:   "bottleneck",
					ID:    "disk-limit-rops",
					Score: score,
					Res: Resource{
						Typ:           "disk",
						Name:          fmt.Sprintf("disk:%s", dname),
						PropertyName:  "read-ops",
						PropertyValue: mean,
					},
				})
			}
		}

		return issues
	})
}
