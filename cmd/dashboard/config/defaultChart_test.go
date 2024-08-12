package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseChartString(t *testing.T) {
	chartString := "title=onechart,name=onechart,repo=https://chart.onechart.dev,version=0.54.0"
	chart, err := parseDefaultChartString(chartString)
	assert.Nil(t, err)
	assert.Equal(t, "onechart", chart.Title)
	assert.Equal(t, "onechart", chart.Chart.Name)

	chartString = "name=https://github.com/mycompany/onechart.git?sha=xxx&path=/charts/onechart/"
	chart, err = parseDefaultChartString(chartString)
	assert.Nil(t, err)
	assert.Equal(t, "https://github.com/mycompany/onechart.git?sha=xxx&path=/charts/onechart/", chart.Chart.Name)
}
