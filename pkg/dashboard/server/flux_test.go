package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fluxEvents "github.com/fluxcd/pkg/apis/event/v1beta1"
	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_fluxEvent(t *testing.T) {
	notificationsManager := notifications.NewDummyManager()
	gitopsRepos := map[string]*config.GitopsRepoConfig{}

	event := fluxEvents.Event{
		InvolvedObject: corev1.ObjectReference{
			Kind:      "GitRepository",
			Namespace: "gitops-system",
			Name:      "webapp",
		},
		Severity:  "info",
		Timestamp: metav1.Now(),
		Message:   "message",
		Reason:    "reason",
		Metadata: map[string]string{
			"test":     "metadata",
			"revision": "xyz",
		},
		ReportingController: "source-controller",
		ReportingInstance:   "source-controller-xyz",
	}

	body, _ := json.Marshal(event)

	fmt.Println(string(body))

	_, _, err := testPostEndpoint(fluxEvent, func(ctx context.Context) context.Context {
		ctx = context.WithValue(ctx, "notificationsManager", notificationsManager)
		ctx = context.WithValue(ctx, "gitopsRepo", "my/gitops")
		ctx = context.WithValue(ctx, "gitopsRepos", gitopsRepos)
		ctx = context.WithValue(ctx, "store", store.NewTest(encryptionKey, encryptionKeyNew))
		return ctx
	}, "/path", string(body))
	assert.Nil(t, err)
}

func testPostEndpoint(handlerFunc http.HandlerFunc, cn contextFunc, path string, body string) (int, string, error) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	req = req.WithContext(cn(req.Context()))

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlerFunc)
	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	return rr.Code, rr.Body.String(), nil
}

func Test_parseRev(t *testing.T) {
	parsed, _ := parseRev("main/1234567890")
	assert.Equal(t, "1234567890", parsed)

	parsed, _ = parseRev("main@sha1:69b59063470310ebbd88a9156325322a124e55a3")
	assert.Equal(t, "69b59063470310ebbd88a9156325322a124e55a3", parsed)
}
