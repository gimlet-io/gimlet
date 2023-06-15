package config

import (
	"testing"

	"gotest.tools/assert"
)

func Test_UpdateConfig(t *testing.T) {
	toUpdate := &Config{
		JWTSecret:                 "xyz",
		TermsOfServiceFeatureFlag: true,
	}

	new := &Config{
		Host:                      "host",
		JWTSecret:                 "cccc",
		TermsOfServiceFeatureFlag: true,
		PrintAdminToken:           true,
		ReleaseHistorySinceDays:   100,
	}

	updateConfigWhenZeroValue(toUpdate, new)

	assert.Equal(t, toUpdate.Host, "host", "Empty strings should be updated")
	assert.Equal(t, toUpdate.JWTSecret, "xyz", "Set values should not be updated")
	assert.Equal(t, toUpdate.TermsOfServiceFeatureFlag, true, "Equal values should stay equal")
	assert.Equal(t, toUpdate.PrintAdminToken, true, "Booelan fields should be set")
	assert.Equal(t, toUpdate.ReleaseHistorySinceDays, 100, "Int fields should be set")
}

func Test_UpdateConfig_StructFields(t *testing.T) {
	toUpdate := &Config{
		Github: Github{
			AppID: "appid",
		},
	}

	new := &Config{
		Github: Github{
			ClientID: "clientid",
		},
	}

	updateConfigWhenZeroValue(toUpdate, new)

	assert.Equal(t, toUpdate.Github.AppID, "appid", "Existing values should be preserved")
	assert.Equal(t, toUpdate.Github.ClientID, "clientid", "Zero values should be updated")
}
