package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_redirectPath(t *testing.T) {
	path := redirectPath("")
	assert.Equal(t, "/", path)

	path = redirectPath("15338405274591342306&https://localhost/auth")
	assert.Equal(t, "/", path)

	path = redirectPath("15338405274591342306&https://localhost/auth&redirect=/repo/gimlet/gimlet-io")
	assert.Equal(t, "/repo/gimlet/gimlet-io", path)
}
