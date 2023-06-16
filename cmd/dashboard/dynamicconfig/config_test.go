package dynamicconfig

import (
	"testing"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"gotest.tools/assert"
)

func Test_UpdateConfig(t *testing.T) {
	toUpdate := &DynamicConfig{
		DummyString: "xyz",
		DummyBool:   true,
	}

	new := &DynamicConfig{
		DummyString:  "cccc",
		DummyString2: "host",
		DummyBool:    true,
		DummyBool2:   true,
		DummyInt:     100,
	}

	updateConfigWhenZeroValue(toUpdate, new)

	assert.Equal(t, toUpdate.DummyString, "xyz", "Set values should not be updated")
	assert.Equal(t, toUpdate.DummyString2, "host", "Empty strings should be updated")
	assert.Equal(t, toUpdate.DummyBool, true, "Equal values should stay equal")
	assert.Equal(t, toUpdate.DummyBool2, true, "Booelan fields should be set")
	assert.Equal(t, toUpdate.DummyInt, 100, "Int fields should be set")
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
