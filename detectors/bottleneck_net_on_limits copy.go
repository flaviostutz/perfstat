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

		for nname, nm := range ActiveStats.NetStats.NICs {

			//TODO add Dropped packets as a catalyser for this analysis?

			//NET LIMIT ON SENT BPS
			score, mean := upperRateBoundaries(&nm.BytesSent, fromLimit, to, opt, 10000.0, [2]float64{0.8, 0.9})
			if score > 0 {
				logrus.Tracef("net-limit-sbps bps=%.2f criticityScore=%.2f", mean, score)

				//get hungry processes
				related := make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopNetByteRate(false) {
					if len(related) >= 3 {
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
					r := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("pid:%d", proc.Pid),
						PropertyName:  "net-sent-bps",
						PropertyValue: rate,
					}
					related = append(related, r)
				}

				issues = append(issues, Issue{
					Typ:   "bottleneck",
					ID:    "net-limit-sbps",
					Score: score,
					Res: Resource{
						Typ:           "net",
						Name:          fmt.Sprintf("nic:%s", nname),
						PropertyName:  "send-bps",
						PropertyValue: mean,
					},
					Related: related,
				})
			}

			//NET LIMIT ON RECV BPS
			score, mean = upperRateBoundaries(&nm.BytesRecv, fromLimit, to, opt, 10000.0, [2]float64{0.8, 0.9})
			if score > 0 {
				logrus.Tracef("net-limit-rbps bps=%.2f criticityScore=%.2f", mean, score)

				//get hungry processes
				related := make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopNetByteRate(true) {
					if len(related) >= 3 {
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
					r := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("pid:%d", proc.Pid),
						PropertyName:  "net-recv-bps",
						PropertyValue: rate,
					}
					related = append(related, r)
				}

				issues = append(issues, Issue{
					Typ:   "bottleneck",
					ID:    "net-limit-rbps",
					Score: score,
					Res: Resource{
						Typ:           "net",
						Name:          fmt.Sprintf("nic:%s", nname),
						PropertyName:  "recv-bps",
						PropertyValue: mean,
					},
					Related: related,
				})
			}

			//NET LIMIT ON SENT OPS
			score, mean = upperRateBoundaries(&nm.PacketsSent, fromLimit, to, opt, 20.0, [2]float64{0.8, 0.9})
			if score > 0 {
				logrus.Tracef("net-limit-sops ops=%.2f criticityScore=%.2f", mean, score)

				//get hungry processes
				related := make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopNetPacketRate(false) {
					if len(related) >= 3 {
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
					r := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("pid:%d", proc.Pid),
						PropertyName:  "net-sent-pps",
						PropertyValue: rate,
					}
					related = append(related, r)
				}

				issues = append(issues, Issue{
					Typ:   "bottleneck",
					ID:    "net-limit-spps",
					Score: score,
					Res: Resource{
						Typ:           "net",
						Name:          fmt.Sprintf("nic:%s", nname),
						PropertyName:  "sent-pps",
						PropertyValue: mean,
					},
				})
			}

			//NET LIMIT ON RECV PPS
			score, mean = upperRateBoundaries(&nm.PacketsRecv, fromLimit, to, opt, 20.0, [2]float64{0.8, 0.9})
			if score > 0 {
				logrus.Tracef("net-limit-rpps ops=%.2f criticityScore=%.2f", mean, score)

				//get hungry processes
				related := make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopNetPacketRate(true) {
					if len(related) >= 3 {
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
					r := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("pid:%d", proc.Pid),
						PropertyName:  "net-recv-pps",
						PropertyValue: rate,
					}
					related = append(related, r)
				}

				issues = append(issues, Issue{
					Typ:   "bottleneck",
					ID:    "net-limit-rpps",
					Score: score,
					Res: Resource{
						Typ:           "net",
						Name:          fmt.Sprintf("nic:%s", nname),
						PropertyName:  "recv-pps",
						PropertyValue: mean,
					},
				})
			}
		}

		return issues
	})
}
