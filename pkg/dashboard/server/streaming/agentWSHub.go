// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package streaming

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type WSMessage struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type PodLogWSMessage struct {
	Message string `json:"message"`
	Pod     string `json:"pod"`
}

type ImageBuildStatusWSMessage struct {
	BuildId  string `json:"buildId"`
	Status   string `json:"status"`
	LogLine  string `json:"logLine"`
	ClientId string `json:"clientId"`
}

// Client is a middleman between the websocket connection and the hub.
type AgentWSClient struct {
	hub *AgentWSHub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.

// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *AgentWSClient) readPump() {
	defer func() {
		c.hub.Unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		var wsMessage WSMessage
		err = json.Unmarshal(message, &wsMessage)
		if err != nil {
			log.Errorf("could not decode ws message from agent")
		}

		if wsMessage.Type == "tick" {
			continue
		}

		if wsMessage.Type == "logs" {
			var podLogWSMessage PodLogWSMessage
			err = json.Unmarshal([]byte(wsMessage.Payload), &podLogWSMessage)
			if err != nil {
				log.Errorf("could not decode podlog ws message from agent")
			}

			jsonString, _ := json.Marshal(PodLogsEvent{
				StreamingEvent: StreamingEvent{Event: PodLogsEventString},
				Pod:            podLogWSMessage.Pod,
				PodLogs:        podLogWSMessage.Message,
			})
			c.hub.ClientHub.Broadcast <- jsonString
		}

		if wsMessage.Type == "imageBuildLogs" {
			var imageBuildStatus ImageBuildStatusWSMessage
			err = json.Unmarshal([]byte(wsMessage.Payload), &imageBuildStatus)
			if err != nil {
				log.Errorf("could not decode image build log ws message from agent")
			}

			jsonString, _ := json.Marshal(ImageBuildLogEvent{
				StreamingEvent: StreamingEvent{Event: ImageBuildLogEventString},
				BuildId:        imageBuildStatus.BuildId,
				Status:         imageBuildStatus.Status,
				LogLine:        imageBuildStatus.LogLine,
			})
			c.hub.ClientHub.Send <- &ClientMessage{
				ClientId: imageBuildStatus.ClientId,
				Message:  jsonString,
			}

			if imageBuildStatus.Status == "success" {
				go createDeployRequest(
					deployRequest,
					magicEnv,
					store,
					tag,
					user.Login,
					clientHub,
					user.Login,
					string(imageBuildId),
					gitRepoCache,
					magicEnv.Name,
				)
			}
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *AgentWSClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs handles websocket requests from the peer.
func ServeAgentWs(hub *AgentWSHub, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &AgentWSClient{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.Register <- client

	go client.writePump()
	go client.readPump()
}

// ClientHub maintains the set of active clients and broadcasts messages to the
// clients.
type AgentWSHub struct {
	// Registered clients.
	AgentWSClients map[*AgentWSClient]bool

	// Register requests from the clients.
	Register chan *AgentWSClient

	// Unregister requests from clients.
	Unregister chan *AgentWSClient

	ClientHub *ClientHub
}

func NewAgentWSHub(clientHub ClientHub) *AgentWSHub {
	return &AgentWSHub{
		Register:       make(chan *AgentWSClient),
		Unregister:     make(chan *AgentWSClient),
		AgentWSClients: make(map[*AgentWSClient]bool),
		ClientHub:      &clientHub,
	}
}

func (h *AgentWSHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.AgentWSClients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.AgentWSClients[client]; ok {
				delete(h.AgentWSClients, client)
				close(client.send)
			}
		}
	}
}

func createDeployRequest(
	deployRequest dx.MagicDeployRequest,
	builtInEnv *model.Environment,
	store *store.Store,
	tag string,
	triggeredBy string,
	clientHub *ClientHub,
	userLogin string,
	imageBuildId string,
	gitRepoCache *nativeGit.RepoCache,
	builtInEnvName string,
) {
	envConfig, _ := defaultEnvConfig(
		deployRequest.Owner, deployRequest.Repo, deployRequest.Sha, builtInEnvName,
		gitRepoCache,
	)

	artifact, err := createDummyArtifact(
		deployRequest.Owner, deployRequest.Repo, deployRequest.Sha,
		builtInEnv.Name,
		store,
		"127.0.0.1:32447/"+deployRequest.Repo,
		tag,
		envConfig,
	)
	if err != nil {
		logrus.Errorf("cannot create artifact: %s", err)
		streamArtifactCreatedEvent(clientHub, userLogin, imageBuildId, "error", "")
		return
	}

	releaseRequestStr, err := json.Marshal(dx.ReleaseRequest{
		Env:         builtInEnv.Name,
		App:         deployRequest.Repo,
		ArtifactID:  artifact.ID,
		TriggeredBy: triggeredBy,
	})
	if err != nil {
		logrus.Errorf("%s - cannot serialize release request: %s", http.StatusText(http.StatusInternalServerError), err)
		streamArtifactCreatedEvent(clientHub, userLogin, imageBuildId, "error", "")
		return
	}

	artifactEvent, err := store.Artifact(artifact.ID)
	if err != nil {
		logrus.Errorf("%s - cannot find artifact with id %s", http.StatusText(http.StatusNotFound), artifact.ID)
		streamArtifactCreatedEvent(clientHub, userLogin, imageBuildId, "error", "")
		return
	}
	event, err := store.CreateEvent(&model.Event{
		Type:       model.ReleaseRequestedEvent,
		Blob:       string(releaseRequestStr),
		Repository: artifactEvent.Repository,
	})
	if err != nil {
		logrus.Errorf("%s - cannot save release request: %s", http.StatusText(http.StatusInternalServerError), err)
		streamArtifactCreatedEvent(clientHub, userLogin, imageBuildId, "error", "")
		return
	}

	streamArtifactCreatedEvent(clientHub, userLogin, imageBuildId, "created", event.ID)
}

func defaultEnvConfig(
	owner string, repoName string, sha string, env string,
	gitRepoCache *nativeGit.RepoCache,

) (*dx.Manifest, error) {
	repo, err := gitRepoCache.InstanceForRead(fmt.Sprintf("%s/%s", owner, repoName))
	if err != nil {
		return nil, fmt.Errorf("cannot get repo: %s", err)
	}

	files, err := nativeGit.RemoteFolderOnHashWithoutCheckout(repo, sha, ".gimlet")
	if err != nil {
		if strings.Contains(err.Error(), "directory not found") {
			return nil, nil
		} else {
			return nil, fmt.Errorf("cannot list files in .gimlet/: %s", err)
		}
	}

	for _, content := range files {
		var envConfig dx.Manifest
		err = yaml.Unmarshal([]byte(content), &envConfig)
		if err != nil {
			logrus.Warnf("cannot parse env config string: %s", err)
			continue
		}
		if envConfig.Env == env && envConfig.App == repoName {
			return &envConfig, nil
		}
	}

	return nil, nil
}

func createDummyArtifact(
	owner, repo, sha string,
	env string,
	store *store.Store,
	image, tag string,
	envConfig *dx.Manifest,
) (*dx.Artifact, error) {

	if envConfig == nil {
		envConfig = &dx.Manifest{
			App:       repo,
			Namespace: "default",
			Env:       env,
			Chart: dx.Chart{
				Name:       config.DEFAULT_CHART_NAME,
				Repository: config.DEFAULT_CHART_REPO,
				Version:    config.DEFAULT_CHART_VERSION,
			},
			Values: map[string]interface{}{
				"containerPort": 80,
				"gitRepository": owner + "/" + repo,
				"gitSha":        sha,
				"image": map[string]interface{}{
					"repository": image,
					"tag":        tag,
					"pullPolicy": "Always",
				},
				"resources": map[string]interface{}{
					"ignore": true,
				},
			},
		}
	}

	artifact := dx.Artifact{
		ID:           fmt.Sprintf("%s-%s", owner+"/"+repo, uuid.New().String()),
		Created:      time.Now().Unix(),
		Fake:         true,
		Environments: []*dx.Manifest{envConfig},
		Version: dx.Version{
			RepositoryName: owner + "/" + repo,
			SHA:            sha,
			Created:        time.Now().Unix(),
			Branch:         "main",
			AuthorName:     "TODO",
			AuthorEmail:    "TODO",
			CommitterName:  "TODO",
			CommitterEmail: "TODO",
			Message:        "TODO",
			URL:            "TODO",
		},
		Vars: map[string]string{
			"SHA": sha,
		},
	}

	event, err := model.ToEvent(artifact)
	if err != nil {
		return nil, fmt.Errorf("cannot convert to artifact model: %s", err)
	}

	_, err = store.CreateEvent(event)
	if err != nil {
		return nil, fmt.Errorf("cannot save artifact: %s", err)
	}

	return &artifact, nil
}

func streamArtifactCreatedEvent(clientHub *ClientHub, userLogin string, imageBuildId string, status string, trackingId string) {
	jsonString, _ := json.Marshal(ArtifactCreatedEvent{
		StreamingEvent: StreamingEvent{Event: ArtifactCreatedEventString},
		BuildId:        imageBuildId,
		TrackingId:     trackingId,
	})
	clientHub.Send <- &ClientMessage{
		ClientId: userLogin,
		Message:  jsonString,
	}
}
