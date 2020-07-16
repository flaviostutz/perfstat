package detectors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCriticityScore(t *testing.T) {
	v := criticityScore(0.3, [2]float64{0.3, 0.6})
	assert.InDeltaf(t, float64(0), v, 0.01, "")

	v = criticityScore(0, [2]float64{0.3, 0.6})
	assert.InDeltaf(t, float64(0), v, 0.01, "")

	v = criticityScore(0.5, [2]float64{0.3, 0.7})
	assert.InDeltaf(t, 0.5, v, 0.01, "")

	v = criticityScore(0.85, [2]float64{0.8, 0.9})
	assert.InDeltaf(t, 0.5, v, 0.01, "")

	v = criticityScore(1.0, [2]float64{0.95, 1.0})
	assert.InDeltaf(t, 1.0, v, 0.01, "")

	v = criticityScore(0.975, [2]float64{0.95, 1.0})
	assert.InDeltaf(t, 0.5, v, 0.01, "")

	v = criticityScore(-1, [2]float64{0.5, 0.9})
	assert.InDeltaf(t, float64(0), v, 0.01, "")

	v = criticityScore(111, [2]float64{0.5, 0.9})
	assert.InDeltaf(t, float64(1), v, 0.01, "")

	v = criticityScore(0.5, [2]float64{0.4, 0.9})
	assert.InDeltaf(t, 0.2, v, 0.01, "")

	v = criticityScore(0.7, [2]float64{0.4, 0.9})
	assert.InDeltaf(t, 0.6, v, 0.01, "")

}
