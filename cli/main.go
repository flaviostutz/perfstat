package main

import (
	"fmt"
	"time"

	"github.com/flaviostutz/perfstat"
	"github.com/flaviostutz/perfstat/detectors"
	"github.com/sirupsen/logrus"
)

func main() {

	opt := detectors.NewOptions()
	opt.CPULoadAvgDuration = 10 * time.Second
	opt.IORateLoadDuration = 10 * time.Second
	opt.IOLimitsSpan = 30 * time.Second
	opt.MemAvgDuration = 30 * time.Second
	opt.MemLeakDuration = 30 * time.Second

	p := perfstat.Start(opt)
	p.SetLogLevel(logrus.DebugLevel)
	time.Sleep(6 * time.Second)

	fmt.Printf("PERFSTAT IS ALIVE!")

	for {
		fmt.Printf("\n\n\n")
		fmt.Printf("BOTTLENECK -> [SCORE: %.0f]  CPU: %.0f  DISK: %.0f  MEM: %.0f  NET: %.0f\n", p.Score("bottleneck", "")*100, p.Score("bottleneck", "cpu.*")*100, p.Score("bottleneck", "disk.*")*100, p.Score("bottleneck", "mem.*")*100, p.Score("bottleneck", "net.*")*100)
		fmt.Printf("RISK       -> [SCORE: %.0f]  CPU: %.0f  DISK: %.0f  MEM: %.0f  NET: %.0f\n", p.Score("risk", "")*100, p.Score("risk", "cpu.*")*100, p.Score("risk", "disk.*")*100, p.Score("risk", "mem.*")*100, p.Score("risk", "net.*")*100)
		fmt.Printf("\n\n")
		crit := p.TopCriticity(0.01, "", "", true)
		for _, is := range crit {
			// fmt.Printf(">>>> %s\n", is.String())
			fmt.Printf("[ %.2f %s ] (%s %s=%.2f) (%s)\n", is.Score, is.ID, is.Res.Name, is.Res.PropertyName, is.Res.PropertyValue, is.Typ)
			for _, r := range is.Related {
				fmt.Printf("    > RELATED: %s [%s=%.2f]\n", r.Name, r.PropertyName, r.PropertyValue)
			}
		}
		time.Sleep(5 * time.Second)
	}

}
