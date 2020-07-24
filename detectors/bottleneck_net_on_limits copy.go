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

		for nname, nm := range ActiveStats.NetStats.NICs {

			//TODO add Dropped packets as a catalyser for this analysis?

			//NET LIMIT ON SENT BPS
			r := DetectionResult{
				Typ:  "bottleneck",
				ID:   "net-limit-sbps",
				When: time.Now(),
			}

			score, mean := upperRateBoundaries(&nm.BytesSent, fromLimit, to, opt, 10000.0, [2]float64{0.8, 0.9})
			r.Score = score

			r.Res = Resource{
				Typ:           "net",
				Name:          fmt.Sprintf("nic:%s", nname),
				PropertyName:  "send-bps",
				PropertyValue: mean,
			}

			if r.Score > 0 {

				//get hungry processes
				r.Related = make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopNetByteRate(false) {
					if len(r.Related) >= 3 {
						break
					}
					rate, ok := proc.NetIOCounters[nname].BytesSent.Rate(opt.IORateLoadDuration)
					if !ok {
						logrus.Warnf("Couldn't get process net sent bps for net on limits analysis")
						continue
					}
					if rate < 1000 {
						break
					}
					res := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("pid:%d", proc.Pid),
						PropertyName:  "net-sent-bps",
						PropertyValue: rate,
					}
					r.Related = append(r.Related, res)
				}
			}
			issues = append(issues, r)

			//NET LIMIT ON RECV BPS
			r = DetectionResult{
				Typ:  "bottleneck",
				ID:   "net-limit-rbps",
				When: time.Now(),
			}

			score, mean = upperRateBoundaries(&nm.BytesRecv, fromLimit, to, opt, 10000.0, [2]float64{0.8, 0.9})
			r.Score = score

			r.Res = Resource{
				Typ:           "net",
				Name:          fmt.Sprintf("nic:%s", nname),
				PropertyName:  "recv-bps",
				PropertyValue: mean,
			}

			if r.Score > 0 {
				//get hungry processes
				r.Related = make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopNetByteRate(true) {
					if len(r.Related) >= 3 {
						break
					}
					rate, ok := proc.NetIOCounters[nname].BytesRecv.Rate(opt.IORateLoadDuration)
					if !ok {
						logrus.Warnf("Couldn't get process recv bps for net on limits analysis")
						continue
					}
					if rate < 1000 {
						break
					}
					res := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("pid:%d", proc.Pid),
						PropertyName:  "net-recv-bps",
						PropertyValue: rate,
					}
					r.Related = append(r.Related, res)
				}
			}
			issues = append(issues, r)

			//NET LIMIT ON SENT OPS
			r = DetectionResult{
				Typ:  "bottleneck",
				ID:   "net-limit-spps",
				When: time.Now(),
			}

			score, mean = upperRateBoundaries(&nm.PacketsSent, fromLimit, to, opt, 20.0, [2]float64{0.8, 0.9})
			r.Score = score

			r.Res = Resource{
				Typ:           "net",
				Name:          fmt.Sprintf("nic:%s", nname),
				PropertyName:  "sent-pps",
				PropertyValue: mean,
			}

			if r.Score > 0 {
				//get hungry processes
				r.Related = make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopNetPacketRate(false) {
					if len(r.Related) >= 3 {
						break
					}
					rate, ok := proc.NetIOCounters[nname].PacketsSent.Rate(opt.IORateLoadDuration)
					if !ok {
						logrus.Warnf("Couldn't get process sent pps for net on limits analysis")
						continue
					}
					if rate < 5 {
						break
					}
					res := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("pid:%d", proc.Pid),
						PropertyName:  "net-sent-pps",
						PropertyValue: rate,
					}
					r.Related = append(r.Related, res)
				}
			}
			issues = append(issues, r)

			//NET LIMIT ON RECV PPS
			r = DetectionResult{
				Typ:  "bottleneck",
				ID:   "net-limit-rpps",
				When: time.Now(),
			}

			score, mean = upperRateBoundaries(&nm.PacketsRecv, fromLimit, to, opt, 10000.0, [2]float64{0.8, 0.9})
			r.Score = score

			r.Res = Resource{
				Typ:           "net",
				Name:          fmt.Sprintf("nic:%s", nname),
				PropertyName:  "recv-pps",
				PropertyValue: mean,
			}

			if r.Score > 0 {
				//get hungry processes
				r.Related = make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopNetPacketRate(true) {
					if len(r.Related) >= 3 {
						break
					}
					rate, ok := proc.NetIOCounters[nname].PacketsRecv.Rate(opt.IORateLoadDuration)
					if !ok {
						logrus.Warnf("Couldn't get process sent pps for net on limits analysis")
						continue
					}
					if rate < 5 {
						break
					}
					res := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("pid:%d", proc.Pid),
						PropertyName:  "net-recv-pps",
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
