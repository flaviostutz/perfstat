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
	build(opt Option, ps *perfstat.Perfstat) (container.Option, error)
	update(opt Option, ps *perfstat.Perfstat, paused bool) error
	onEvent(evt *terminalapi.Keyboard)
}

var (
	opt                Option
	rootc              *container.Container
	ps                 *perfstat.Perfstat
	curScreenCtxCancel context.CancelFunc
	curScreen          screen
	paused             bool
)

func main() {

	flag.Float64Var(&opt.freq, "freq", 1.0, "Analysis frequency. Changes data capture and display refresh frequency. Higher consumes more CPU. Defaults to 1 Hz")
	flag.Float64Var(&opt.sensibility, "sensibility", 6.0, "Lower values (ex.: 0.2) means larger timespan in analysis, leading to more accurate results but slower responses. Higher values (ex.: 5) means short time analysis but may lead to false positives. Defaults to 1.0 which means detecting a continuous 100% CPU in 30s")
	flag.Parse()

	if opt.freq > 20 || opt.freq < 0.05 {
		panic("--freq must be between 0.05 and 20")
	}

	if opt.sensibility > 30 || opt.sensibility < 0.01 {
		panic("--sensibility must be between 0.01 and 30")
	}

	logrus.Debugf("Initializing Perfstat engine")
	opt2 := detectors.NewOptions()
	dur := time.Duration(30/opt.sensibility) * time.Second
	opt2.CPULoadAvgDuration = dur
	opt2.IORateLoadDuration = dur
	opt2.IOLimitsSpan = dur
	opt2.MemAvgDuration = dur
	opt2.MemLeakDuration = (dur * 10)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ps = perfstat.Start(ctx, opt2)
	ps.SetLogLevel(logrus.DebugLevel)
	// time.Sleep(6 * time.Second)

	logrus.Debugf("Initializing UI...")

	t, err := termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	if err != nil {
		panic(err)
	}

	rootc, err = container.New(t, container.ID("root"))
	if err != nil {
		panic(err)
	}

	evtHandler := func(k *terminalapi.Keyboard) {
		if curScreen != nil {
			curScreen.onEvent(k)
		}
		//exit
		if k.Key == keyboard.KeyCtrlC || k.Key == 113 {
			cancel()
			//pause/unpause
		} else if k.Key == 80 || k.Key == 112 {
			paused = !paused
		} else if k.Key == keyboard.KeyEsc {
			showScreen(&home{})
		}
	}

	controller, err := termdash.NewController(t, rootc, termdash.KeyboardSubscriber(evtHandler))
	if err != nil {
		panic(err)
	}

	showScreen(&home{})

	paused = false

	ticker := time.NewTicker(time.Duration(opt.freq) * time.Second).C
	for {
		select {

		case <-ctx.Done():
			//this is used because when using controller some "double closing" arises
			//should be removed after fixing termdash (https://github.com/mum4k/termdash/issues/241)
			defer func() {
				recover()
			}()
			t.Close()

		case <-ticker:
			err := curScreen.update(opt, ps, paused)
			if err != nil {
				return
			}
			err = controller.Redraw()
		}
	}
}

func showScreen(s screen) {
	r, err := s.build(opt, ps)
	if err != nil {
		panic(fmt.Sprintf("Error preparing screen %s. err=%s", s, err))
	}
	rootc.Update("root", r)

	curScreen = s
}
