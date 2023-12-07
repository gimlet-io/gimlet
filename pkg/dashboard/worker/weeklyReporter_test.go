package worker

import (
	"math"
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/stretchr/testify/assert"
)

func Test_Percentage(t *testing.T) {
	res := percentageChange(1000, 830)
	assert.Equal(t, float64(-17.0), res)

	res = percentageChange(0, 0)
	assert.True(t, math.IsNaN(res))

	// anything divided by 0 will become infinity
	res = percentageChange(0, 50)
	assert.True(t, math.IsInf(res, 1))
}

func Test_MostTriggerer(t *testing.T) {
	releases := []*dx.Release{
		{
			TriggeredBy: "policy",
		},
		{
			TriggeredBy: "dzsak",
		},
		{
			TriggeredBy: "dzsak",
		},
		{
			TriggeredBy: "dzsak",
		},
	}

	name := mostTriggeredBy(releases)
	assert.Equal(t, "dzsak", name)
}
