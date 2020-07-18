package stats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemBasic(t *testing.T) {
	// logrus.SetLevel(logrus.DebugLevel)
	s := NewMemStats(60*time.Second, 2)
	time.Sleep(5 * time.Second)

	v, ok := s.Total.GetLastValue()
	assert.True(t, ok)
	assert.GreaterOrEqualf(t, v.Value, 1000.0, "")

	v, ok = s.Free.GetLastValue()
	assert.True(t, ok)
	assert.GreaterOrEqualf(t, v.Value, 1000.0, "")

	v, ok = s.Used.GetLastValue()
	assert.True(t, ok)
	assert.GreaterOrEqualf(t, v.Value, 1000.0, "")

	v, ok = s.Available.GetLastValue()
	assert.True(t, ok)
	assert.GreaterOrEqualf(t, v.Value, 1000.0, "")

	s.Stop()
}