package detectors

import (
	"fmt"
	"time"

	"github.com/flaviostutz/signalutils"
	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []Issue {
		to := time.Now()
		from := to.Add(-opt.IOLimitsSpan)

		issues := make([]Issue, 0)

		for dname, dm := range ActiveStats.DiskStats.Disks {

			//TODO add Dropped packets as a catalyser for this analysis?

			//DISK LIMIT ON WRITE BPS
			score, mean := upperRateBoundaries(&dm.WriteBytes, from, to, opt, 100000.0, [2]float64{0.8, 0.9})
			if score > 0 {
				logrus.Tracef("disk-limit-wbps bps=%.2f criticityScore=%.2f", mean, score)
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
				})
			}

			//DISK LIMIT ON READ BPS
			score, mean = upperRateBoundaries(&dm.ReadBytes, from, to, opt, 100000.0, [2]float64{0.8, 0.9})
			if score > 0 {
				logrus.Tracef("disk-limit-rbps bps=%.2f criticityScore=%.2f", mean, score)
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
				})
			}

			//DISK LIMIT ON WRITE OPS
			score, mean = upperRateBoundaries(&dm.WriteCount, from, to, opt, 100000.0, [2]float64{0.8, 0.9})
			if score > 0 {
				logrus.Tracef("disk-limit-wops ops=%.2f criticityScore=%.2f", mean, score)
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
			score, mean = upperRateBoundaries(&dm.ReadCount, from, to, opt, 100000.0, [2]float64{0.8, 0.9})
			if score > 0 {
				logrus.Tracef("disk-limit-rops ops=%.2f criticityScore=%.2f", mean, score)
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

func upperRateBoundaries(tcr *signalutils.TimeseriesCounterRate, from time.Time, to time.Time, opt *Options, meanHigher float64, criticityRange [2]float64) (cscore float64, meanRate float64) {
	ts, ok := tcr.RateOverTime(opt.IORateLoadDuration, opt.IOLimitsSpan)
	if !ok {
		return 0, 0
	}
	sd, m := ts.StdDev(from, to)
	if m > meanHigher {
		stdDev := (sd / m)
		return criticityScore(1.0-stdDev, criticityRange), m
	}
	return 0.0, m
}
