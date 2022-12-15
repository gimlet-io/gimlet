package main

import (
	"testing"

	"gotest.tools/assert"
)

func TestGroupFromUsername(t *testing.T) {
	assert.Equal(t, "2", groupFromUsername("group_2_bot1"))
}
