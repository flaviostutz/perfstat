package detectors

import (
	"fmt"
	"time"
)

func init() {
	RegisterDetector(func(opt *Options) []DetectionResult {

		r := DetectionResult{
			Typ:  "risk",
			ID:   "mem-swap-high",
			When: time.Now(),
		}

		to := time.Now()
		from := to.Add(-opt.MemAvgDuration)

		sin, ok := ActiveStats.MemStats.SwapIn.Rate(opt.MemAvgDuration)
		if !ok {
			r.Message = notEnoughDataMessage(opt.MemAvgDuration)
			r.Score = -1
			return []DetectionResult{r}
		}
		sout, ok := ActiveStats.MemStats.SwapOut.Rate(opt.MemAvgDuration)
		if !ok {
			r.Message = notEnoughDataMessage(opt.MemAvgDuration)
			r.Score = -1
			return []DetectionResult{r}
		}

		r.Res = Resource{
			Typ:           "mem",
			Name:          "swap",
			PropertyName:  "mem-swap-total-bps",
			PropertyValue: sin + sout,
		}

		r.Score = criticityScore(sin+sout, opt.HighSwapBpsRange)

		if r.Score == 0 {
			return []DetectionResult{r}
		}

		//get processes with high swap
		r.Related = make([]Resource, 0)
		for _, proc := range ActiveStats.ProcessStats.TopMemSwap() {
			if len(r.Related) >= 5 {
				break
			}

			swap, ok := proc.MemorySwap.Avg(from, to)
			if !ok {
				continue
			}
			re := Resource{
				Typ:           "process",
				Name:          fmt.Sprintf("%s(%s)[%d]", proc.Cmdline, proc.Name, proc.Pid),
				PropertyName:  "mem-swap-bytes",
				PropertyValue: swap,
			}
			r.Related = append(r.Related, re)
		}

		return []DetectionResult{r}
	})
}
