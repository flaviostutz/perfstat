package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/flaviostutz/perfstat"
	"github.com/flaviostutz/perfstat/detectors"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/sirupsen/logrus"
)

type Option struct {
	freq        float64
	sensibility float64
}

type screen interface {
	build(ctx context.Context, opt Option, ps *perfstat.Perfstat) ([]container.Option, error)
	onEvent(evt *terminalapi.Keyboard)
}

var (
	opt                Option
	rootc              *container.Container
	ps                 *perfstat.Perfstat
	curScreenCtxCancel context.CancelFunc
	curScreen          screen
	mainCtx            context.Context
)

func main() {

	flag.Float64Var(&opt.freq, "freq", 1.0, "Analysis frequency. Changes data capture and display refresh frequency. Higher consumes more CPU. Defaults to 1 Hz")
	flag.Float64Var(&opt.sensibility, "sensibility", 1.0, "Lower values (ex.: 0.2) means larger timespan in analysis, leading to more accurate results but slower responses. Higher values (ex.: 5) means short time analysis but may lead to false positives. Defaults to 1.0")
	flag.Parse()

	if opt.freq > 20 || opt.freq < 0.05 {
		panic("--freq must be between 0.05 and 20")
	}

	if opt.sensibility > 30 || opt.sensibility < 0.01 {
		panic("--sensibility must be between 0.01 and 30")
	}

	logrus.Debugf("Initializing Perfstat engine")
	opt2 := detectors.NewOptions()
	dur := time.Duration(30*opt.sensibility) * time.Second
	opt2.CPULoadAvgDuration = dur
	opt2.IORateLoadDuration = dur
	opt2.IOLimitsSpan = dur
	opt2.MemAvgDuration = dur
	opt2.MemLeakDuration = (dur * 10)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mainCtx = ctx

	ps = perfstat.Start(ctx, opt2)
	ps.SetLogLevel(logrus.DebugLevel)
	// time.Sleep(6 * time.Second)

	logrus.Debugf("Initializing UI...")

	t, err := termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	if err != nil {
		panic(err)
	}
	defer t.Close()

	rootc, err = container.New(t, container.ID("root"))
	if err != nil {
		panic(err)
	}

	evtHandler := func(k *terminalapi.Keyboard) {
		if curScreen != nil {
			curScreen.onEvent(k)
		}
		if k.Key == keyboard.KeyEsc || k.Key == keyboard.KeyCtrlC {
			cancel()
		}
	}

	showScreen(home{})

	err = termdash.Run(ctx,
		t,
		rootc,
		termdash.KeyboardSubscriber(evtHandler),
		termdash.RedrawInterval(time.Duration(1.0/opt.freq)*time.Second))
	if err != nil {
		panic(err)
	}

	// for {
	// 	fmt.Printf("\n\n\n")
	// 	fmt.Printf("BOTTLENECK -> [SCORE: %.0f]  CPU: %.0f  DISK: %.0f  MEM: %.0f  NET: %.0f\n", p.Score("bottleneck", "")*100, p.Score("bottleneck", "cpu.*")*100, p.Score("bottleneck", "disk.*")*100, p.Score("bottleneck", "mem.*")*100, p.Score("bottleneck", "net.*")*100)
	// 	fmt.Printf("RISK       -> [SCORE: %.0f]  CPU: %.0f  DISK: %.0f  MEM: %.0f  NET: %.0f\n", p.Score("risk", "")*100, p.Score("risk", "cpu.*")*100, p.Score("risk", "disk.*")*100, p.Score("risk", "mem.*")*100, p.Score("risk", "net.*")*100)
	// 	fmt.Printf("\n\n")
	// 	crit := ps.TopCriticity(0.01, "", "", true)
	// 	for _, is := range crit {
	// 		// fmt.Printf(">>>> %s\n", is.String())
	// 		fmt.Printf("[ %.2f %s ] (%s %s=%.2f) (%s)\n", is.Score, is.ID, is.Res.Name, is.Res.PropertyName, is.Res.PropertyValue, is.Typ)
	// 		for _, r := range is.Related {
	// 			fmt.Printf("    > RELATED: %s [%s=%.2f]\n", r.Name, r.PropertyName, r.PropertyValue)
	// 		}
	// 	}
	// 	time.Sleep(5 * time.Second)
	// }
}

func showScreen(s screen) {
	//free previous screen resources
	if curScreenCtxCancel != nil {
		curScreenCtxCancel()
	}

	ctxc, cc := context.WithCancel(mainCtx)
	r, err := s.build(ctxc, opt, ps)
	if err != nil {
		panic(fmt.Sprintf("Error preparing screen %s. err=%s", s, err))
	}
	rootc.Update("root", r...)

	curScreen = s
	curScreenCtxCancel = cc
}
