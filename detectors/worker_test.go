package detectors

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkerStepError(t *testing.T) {
	// logrus.SetLevel(logrus.TraceLevel)
	w := StartWorker("test1", func() error {
		time.Sleep(200 * time.Millisecond)
		return fmt.Errorf("Error here")
	}, 5, true)
	time.Sleep(100 * time.Millisecond)
	assert.True(t, w.active)
	time.Sleep(400 * time.Millisecond)
	assert.False(t, w.active)
}

func TestWorkerStepFreq(t *testing.T) {
	// logrus.SetLevel(logrus.TraceLevel)
	w := StartWorker("test1", func() error {
		time.Sleep(15 * time.Millisecond)
		return nil
	}, 5, true)
	time.Sleep(200 * time.Millisecond)
	assert.True(t, w.active)
	time.Sleep(2000 * time.Millisecond)
	assert.InDeltaf(t, 5, w.CurrentFreq, 2, "")
	assert.InDeltaf(t, 15, w.CurrentStepTime.Milliseconds(), 5, "")
	w.Stop()
	time.Sleep(300 * time.Millisecond)
	assert.False(t, w.active)
}
