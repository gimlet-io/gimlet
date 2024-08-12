package config

import (
	"testing"
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
