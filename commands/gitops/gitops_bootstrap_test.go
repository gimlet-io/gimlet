package gitops

import (
	"testing"
)

func Test_parseRepoURL(t *testing.T) {
	host, owner, repo := parseRepoURL("git@github.com:gimlet-io/gimlet-cli.git")
	if host != "github.com" {
		t.Errorf("Must parse host")
	}
	if owner != "gimlet-io" {
		t.Errorf("Must parse owner")
	}
	if repo != "gimlet-cli" {
		t.Errorf("Must parse repo")
	}
}
