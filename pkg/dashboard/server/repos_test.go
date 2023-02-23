package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Difference(t *testing.T) {
	slice1 := []string{"foo", "bar", "hello"}
	slice2 := []string{"foo", "bar"}

	differences := difference(slice1, slice2)
	assert.Equal(t, 1, len(differences))
	assert.Equal(t, "hello", differences[0])

	differences = difference(slice2, slice1)
	assert.Equal(t, 0, len(differences))
}
