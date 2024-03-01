package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuiltInEnvFlag(t *testing.T) {
	c := &Config{}
	defaults(c)
	flag := c.BuiltinEnvFeatureFlag()
	if !flag {
		t.Errorf("FEATURE_BUILT_IN_ENV should default to true")
	}

	c = &Config{
		BuiltinEnvFeatureFlagString: "false",
	}
	defaults(c)
	flag = c.BuiltinEnvFeatureFlag()
	if flag {
		t.Errorf("FEATURE_BUILT_IN_ENV should accept values")
	}

	c = &Config{
		BuiltinEnvFeatureFlagString: "true",
	}
	defaults(c)
	flag = c.BuiltinEnvFeatureFlag()
	if !flag {
		t.Errorf("FEATURE_BUILT_IN_ENV should accept values")
	}

	c = &Config{
		BuiltinEnvFeatureFlagString: "not a boolean",
	}
	defaults(c)
	flag = c.BuiltinEnvFeatureFlag()
	if !flag {
		t.Errorf("FEATURE_BUILT_IN_ENV should default to true in case of parse errors")
	}
}

func TestParseChartString(t *testing.T) {
	chartString := "name=onechart,repo=https://chart.onechart.dev,version=0.54.0"
	chart, err := parseChartString(chartString)
	assert.Nil(t, err)
	assert.Equal(t, "onechart", chart.Name)

	chart, _ = parseChartString("")
	assert.Nil(t, chart)

	chartString = "name=https://github.com/mycompany/onechart.git?sha=xxx&path=/charts/onechart/"
	chart, err = parseChartString(chartString)
	assert.Nil(t, err)
	assert.Equal(t, "https://github.com/mycompany/onechart.git?sha=xxx&path=/charts/onechart/", chart.Name)

}
