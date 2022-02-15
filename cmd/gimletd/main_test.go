package main

import (
	"testing"

	"github.com/gimlet-io/gimlet-cli/cmd/gimletd/config"
)

func TestParseChannelMapping(t *testing.T) {
	config := &config.Config{
		Notifications: config.Notifications{
			ChannelMapping: "staging=my-team,prod=another-team",
		},
	}

	testChannelMap := parseChannelMap(config)

	assertEqual(t, testChannelMap["staging"], "my-team")
	assertEqual(t, testChannelMap["prod"], "another-team")
}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}
