package detectors

import (
	"time"

	"github.com/sirupsen/logrus"
)

type Worker struct {
	maxFreq         int
	ticker          *time.Ticker
	done            chan (bool)
	step            StepFunc
	stopOnErr       bool
	name            string
	active          bool
	CurrentFreq     float64
	CurrentStepTime time.Duration
}

type StepFunc func() error

func StartWorker(name string, step StepFunc, maxFreq int, stopOnErr bool) *Worker {
	c := &Worker{
		name:      name,
		maxFreq:   maxFreq,
		done:      make(chan bool),
		ticker:    time.NewTicker(time.Duration((int(time.Second) / maxFreq))),
		step:      step,
		stopOnErr: stopOnErr,
		active:    false,
	}
	logrus.Tracef("Starting goroutine for %s", name)
	go c.run()
	return c
}

func (c *Worker) Stop() {
	c.done <- true
}

func (c *Worker) run() {
	c.active = true
	for {
		loopStart := time.Now()
		select {
		case <-c.done:
			c.active = false
			return
		case <-c.ticker.C:
			stepStart := time.Now()
			err := c.step()
			c.CurrentStepTime = time.Since(stepStart)
			loopElapsed := time.Since(loopStart)
			c.CurrentFreq = float64(1) / loopElapsed.Seconds()
			logrus.Debugf("%s: STEP time=%d ms; loop freq=%.2f", c.name, c.CurrentStepTime.Milliseconds(), c.CurrentFreq)
			if err != nil {
				logrus.Infof("%s: STEP err=%s", c.name, err)
				if c.stopOnErr {
					c.active = false
					return
				}
			}
		}
	}
}
