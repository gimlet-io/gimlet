package server

import (
	"encoding/base32"
	"fmt"
	"testing"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/gorilla/securecookie"
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

func Test_TriggerConcurrentMapWritesError(t *testing.T) {
	stopCh := make(chan struct{})
	defer close(stopCh)

	store := store.New(
		"sqlite",
		"/Users/adam/Desktop/work/gimlet/cmd/dashboard/gimlet-dashboard.sqlite?_pragma=busy_timeout=10000",
		"",
		"",
	)
	dynamicConfig, _ := dynamicconfig.LoadDynamicConfig(store)
	tokenManager := customScm.NewTokenManager(dynamicConfig)
	repoCache, _ := nativeGit.NewRepoCache(
		tokenManager,
		stopCh,
		&config.Config{RepoCachePath: "/tmp/gimlet-dashboard"},
		&dynamicconfig.DynamicConfig{
			Github: config.Github{
				AppID: "app-id",
			},
		},
		nil,
		&model.User{
			Login: "git",
			Secret: base32.StdEncoding.EncodeToString(
				securecookie.GenerateRandomKey(32),
			),
			Admin: false,
		},
	)
	go repoCache.Run()

	for i := 0; i < 1000; i++ {
		repo, _ := repoCache.InstanceForRead("dzsak/friendly-octo-sniffle")
		go func(i int) {
			fmt.Println("Loop repo read - ", i)
			nativeGit.HeadBranch(repo)
			nativeGit.RemoteFolderOnBranchWithoutCheckout(repo, "main", ".gimlet")
		}(i)
	}

	time.Sleep(3 * time.Second)
}
