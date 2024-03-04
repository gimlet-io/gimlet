// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package streaming

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type WSMessage struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type PodLogWSMessage struct {
	Timestamp  string `json:"timestamp"`
	Container  string `json:"container"`
	Message    string `json:"message"`
	Pod        string `json:"pod"`
	Deployment string `json:"deployment"`
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

		if wsMessage.Type == "log" {
			var podLogWSMessage PodLogWSMessage
			err = json.Unmarshal([]byte(wsMessage.Payload), &podLogWSMessage)
			if err != nil {
				log.Errorf("could not decode podlog ws message from agent")
			}

			jsonString, _ := json.Marshal(PodLogsEvent{
				StreamingEvent: StreamingEvent{Event: PodLogsEventString},
				Timestamp:      podLogWSMessage.Timestamp,
				Container:      podLogWSMessage.Container,
				Pod:            podLogWSMessage.Pod,
				Deployment:     podLogWSMessage.Deployment,
				Message:        podLogWSMessage.Message,
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

			if imageBuildStatus.Status != "running" {
				c.hub.successfullImageBuilds <- imageBuildStatus
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

	successfullImageBuilds chan ImageBuildStatusWSMessage
}

func NewAgentWSHub(clientHub ClientHub, successfullImageBuilds chan ImageBuildStatusWSMessage) *AgentWSHub {
	return &AgentWSHub{
		Register:               make(chan *AgentWSClient),
		Unregister:             make(chan *AgentWSClient),
		AgentWSClients:         make(map[*AgentWSClient]bool),
		ClientHub:              &clientHub,
		successfullImageBuilds: successfullImageBuilds,
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
