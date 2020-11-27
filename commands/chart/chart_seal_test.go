package chart

import (
	"github.com/franela/goblin"
	"github.com/gimlet-io/gimlet-cli/commands"
	"io/ioutil"
	"os"
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

			tmpFile, err := ioutil.TempFile("", "dummyStdIn")
			g.Assert(err == nil).IsTrue()
			defer os.Remove(tmpFile.Name())
			_, err = tmpFile.Write([]byte(toSeal))
			g.Assert(err == nil).IsTrue()
			_, err = tmpFile.Seek(0, 0)
			g.Assert(err == nil).IsTrue()
			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }()

			os.Stdin = tmpFile

			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)
		})
	})

}
