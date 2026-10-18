package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 65536
)

type VoiceGuard interface {
	StartTransmission(userID, channelID string) error
	StopTransmission(userID, channelID string) error
	CanTransmit(userID, channelID string) bool
	IsChannelBusy(channelID string) bool
}

type Client struct {
	hub        *Hub
	conn       *websocket.Conn
	send       chan Message
	ID         string
	UserID     string
	rooms      map[string]bool
	voiceGuard VoiceGuard
}

type ClientMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func NewClient(hub *Hub, conn *websocket.Conn, userID string) *Client {
	return &Client{
		hub:        hub,
		conn:       conn,
		send:       make(chan Message, 256),
		ID:         generateClientID(),
		UserID:     userID,
		rooms:      make(map[string]bool),
		voiceGuard: nil,
	}
}

func NewClientWithGuard(hub *Hub, conn *websocket.Conn, userID string, guard VoiceGuard) *Client {
	return &Client{
		hub:        hub,
		conn:       conn,
		send:       make(chan Message, 256),
		ID:         generateClientID(),
		UserID:     userID,
		rooms:      make(map[string]bool),
		voiceGuard: guard,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.UnregisterClient(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var clientMsg ClientMessage
		if err := json.Unmarshal(message, &clientMsg); err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}

		c.handleMessage(clientMsg)
	}
}

func (c *Client) WritePump() {
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
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			jsonBytes, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}

			if _, err := w.Write(jsonBytes); err != nil {
				return
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

func (c *Client) handleMessage(msg ClientMessage) {
	switch msg.Type {
	case "join_room":
		var payload struct {
			Room string `json:"room"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			c.hub.JoinRoom(c, payload.Room)
		}

	case "leave_room":
		var payload struct {
			Room string `json:"room"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			c.hub.LeaveRoom(c, payload.Room)
		}

	case "audio":
		var payload struct {
			Room   string          `json:"room"`
			Audio  json.RawMessage `json:"audio"`
			IsPTT  bool            `json:"is_ptt"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			if c.voiceGuard != nil && payload.IsPTT {
				if err := c.voiceGuard.StartTransmission(c.UserID, payload.Room); err != nil {
					log.Printf("PTT start rejected for user %s in room %s: %v", c.UserID, payload.Room, err)
					return
				}
			} else if c.voiceGuard != nil && !payload.IsPTT {
				c.voiceGuard.StopTransmission(c.UserID, payload.Room)
			}
			audioPayload, _ := json.Marshal(map[string]interface{}{
				"audio":    payload.Audio,
				"is_ptt":   payload.IsPTT,
				"user_id":  c.UserID,
				"sender":   c.ID,
			})
			c.hub.BroadcastToRoom(payload.Room, Message{
				Type:    "audio",
				Payload: audioPayload,
				Sender:  c,
			})
		}

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

func generateClientID() string {
	return time.Now().Format("20060102150405.000000000")
}
