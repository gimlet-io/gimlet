package chart

import (
	"github.com/franela/goblin"
	"github.com/gimlet-io/gimlet-cli/commands"
	"strings"
	"testing"
)

func Test_seal(t *testing.T) {
	args := strings.Split("gimlet chart seal", " ")
	args = append(args, "-f", "-")
	args = append(args, "-p", ".sealedSecrets")

	g := goblin.Goblin(t)

	g.Describe("gimlet chart seal", func() {
		g.It("Should seal stdin", func() {

			const toSeal = `
key: value
another: one

sealedSecrets:
  secret1: value1
  secret2: value2
`

			err := commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)
		})
	})

}
