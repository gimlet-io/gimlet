package dynamicconfig

import (
	"testing"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"gotest.tools/assert"
)

func Test_UpdateConfig(t *testing.T) {
	toUpdate := &DynamicConfig{
		AdminKey: "admin-key",
		Github: config.Github{
			Debug: true,
		},
	}

	new := &DynamicConfig{
		AdminKey:  "new-admin-key",
		JWTSecret: "new-jwt-secret",
		Github: config.Github{
			Debug:      true,
			SkipVerify: true,
		},
	}

	updateConfigWhenZeroValue(toUpdate, new)

	assert.Equal(t, toUpdate.AdminKey, "admin-key", "Set values should not be updated")
	assert.Equal(t, toUpdate.JWTSecret, "new-jwt-secret", "Empty strings should be updated")
	assert.Equal(t, toUpdate.Github.Debug, true, "Equal values should stay equal")
	assert.Equal(t, toUpdate.Github.SkipVerify, true, "Booelan fields should be set")
}

func Test_UpdateConfig_StructFields(t *testing.T) {
	toUpdate := &DynamicConfig{
		Github: config.Github{
			AppID: "appid",
		},
	}

	new := &DynamicConfig{
		Github: config.Github{
			ClientID: "clientid",
		},
	}

	updateConfigWhenZeroValue(toUpdate, new)

	assert.Equal(t, toUpdate.Github.AppID, "appid", "Existing values should be preserved")
	assert.Equal(t, toUpdate.Github.ClientID, "clientid", "Zero values should be updated")
}
