package detectors

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []DetectionResult {

		to := time.Now()
		fromLimit := to.Add(-opt.IOLimitsSpan)

		issues := make([]DetectionResult, 0)

		for dname, dm := range ActiveStats.DiskStats.Disks {

			//TODO add Dropped packets as a catalyser for this analysis?

			//DISK LIMIT ON WRITE BPS
			score, mean := upperRateBoundaries(&dm.WriteBytes, fromLimit, to, opt, 100000.0, [2]float64{0.8, 0.9})

			r := DetectionResult{
				Typ:   "bottleneck",
				ID:    "disk-limit-wbps",
				When:  time.Now(),
				Score: score,
				Res: Resource{
					Typ:           "disk",
					Name:          fmt.Sprintf("disk:%s", dname),
					PropertyName:  "write-bps",
					PropertyValue: mean,
				},
			}

			if r.Score > 0 {
				//get hungry processes
				r.Related = make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopIOByteRate(false) {
					if len(r.Related) >= 3 {
						break
					}
					rate, ok := proc.IOCounters.WriteBytes.Rate(opt.IORateLoadDuration)
					if !ok {
						logrus.Warnf("Not enough data to calculate process io rate")
						continue
					}
					if rate < 10000 {
						break
					}
					res := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("%s(%s)[%d]", proc.Cmdline, proc.Name, proc.Pid),
						PropertyName:  "disk-write-bps",
						PropertyValue: rate,
					}
					r.Related = append(r.Related, res)
				}
			}
			issues = append(issues, r)

			//DISK LIMIT ON READ BPS
			score, mean = upperRateBoundaries(&dm.ReadBytes, fromLimit, to, opt, 100000.0, [2]float64{0.8, 0.9})

			r = DetectionResult{
				Typ:   "bottleneck",
				ID:    "disk-limit-rbps",
				Score: score,
				When:  time.Now(),
				Res: Resource{
					Typ:           "disk",
					Name:          fmt.Sprintf("disk:%s", dname),
					PropertyName:  "read-bps",
					PropertyValue: mean,
				},
			}

			if r.Score > 0 {
				//get hungry processes
				r.Related = make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopIOByteRate(true) {
					if len(r.Related) >= 3 {
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
					res := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("%s(%s)[%d]", proc.Cmdline, proc.Name, proc.Pid),
						PropertyName:  "disk-read-bps",
						PropertyValue: rate,
					}
					r.Related = append(r.Related, res)
				}
			}
			issues = append(issues, r)

			//DISK LIMIT ON WRITE OPS
			score, mean = upperRateBoundaries(&dm.WriteCount, fromLimit, to, opt, 100000.0, [2]float64{0.8, 0.9})

			r = DetectionResult{
				Typ:   "bottleneck",
				ID:    "disk-limit-wops",
				When:  time.Now(),
				Score: score,
				Res: Resource{
					Typ:           "disk",
					Name:          fmt.Sprintf("disk:%s", dname),
					PropertyName:  "write-ops",
					PropertyValue: mean,
				},
			}

			if r.Score > 0 {
				//get hungry processes
				r.Related = make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopIOOpRate(false) {
					if len(r.Related) >= 3 {
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
					res := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("%s(%s)[%d]", proc.Cmdline, proc.Name, proc.Pid),
						PropertyName:  "disk-write-ops",
						PropertyValue: rate,
					}
					r.Related = append(r.Related, res)
				}
			}
			issues = append(issues, r)

			//DISK LIMIT ON READ OPS
			score, mean = upperRateBoundaries(&dm.ReadCount, fromLimit, to, opt, 100000.0, [2]float64{0.8, 0.9})

			r = DetectionResult{
				Typ:   "bottleneck",
				ID:    "disk-limit-rops",
				When:  time.Now(),
				Score: score,
				Res: Resource{
					Typ:           "disk",
					Name:          fmt.Sprintf("disk:%s", dname),
					PropertyName:  "read-ops",
					PropertyValue: mean,
				},
			}

			if r.Score > 0 {
				//get hungry processes
				r.Related = make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopIOOpRate(true) {
					if len(r.Related) >= 3 {
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
					res := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("%s(%s)[%d]", proc.Cmdline, proc.Name, proc.Pid),
						PropertyName:  "disk-read-ops",
						PropertyValue: rate,
					}
					r.Related = append(r.Related, res)
				}
			}
			issues = append(issues, r)
		}

		return issues
	})
}
