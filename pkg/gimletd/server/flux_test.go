package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fluxcd/pkg/runtime/events"
	"github.com/gimlet-io/gimlet-cli/cmd/gimletd/config"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/store"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_fluxEvent(t *testing.T) {
	notificationsManager := notifications.NewDummyManager()
	gitopsRepos := []*config.GitopsRepoConfig{}
	config := config.Config{}
	eventSinkHub := streaming.NewEventSinkHub(&config)

	event := events.Event{
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
		ctx = context.WithValue(ctx, "store", store.NewTest())
		ctx = context.WithValue(ctx, "eventSinkHub", eventSinkHub)
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
