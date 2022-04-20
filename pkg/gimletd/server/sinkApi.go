package server

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/server/streaming"
	"github.com/sirupsen/logrus"
)

func register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Streaming not supported"))
		return
	}

	io.WriteString(w, ": ping\n\n")
	flusher.Flush()

	logrus.Debugf("sink connected")

	eventChannel := make(chan interface{}, 10)
	defer func() {
		<-r.Context().Done()
		close(eventChannel)
		logrus.Debugf("sink disconnected")
	}()

	a := &streaming.EventSink{EventChannel: eventChannel}

	hub, _ := r.Context().Value("eventSinkHub").(*streaming.EventSinkHub)
	hub.Register <- a

	for {
		select {
		case <-r.Context().Done():
			hub.Unregister <- a
			return
		case <-time.After(time.Second * 30):
			io.WriteString(w, ": ping\n\n")
			flusher.Flush()
		case event, ok := <-eventChannel:
			if ok {
				eventString, err := json.Marshal(event)
				if err != nil {
					logrus.Warnf("couldn't marshal event: %s", err)
					continue
				}

				io.WriteString(w, "data: ")
				w.Write(eventString)
				io.WriteString(w, "\n\n")
				flusher.Flush()
			}
		}
	}
}
