package websocket

import (
	"encoding/base64"
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

// BufferedAudio is a replayable audio segment. The concrete AudioBufferService
// lives in internal/voice and cannot be imported here (import cycle), so this
// package only knows this DTO and the AudioBuffer interface (DIP).
type BufferedAudio struct {
	Audio     []byte
	UserID    string
	Sequence  int64
	Timestamp time.Time
}

// AudioBuffer stores recent audio per channel for replay on reconnection (RF06).
type AudioBuffer interface {
	AddMessage(channelID, userID string, audioData []byte) (sequence int64, timestamp time.Time, err error)
	GetMessages(channelID string, sinceSequence int64) []BufferedAudio
}

type Client struct {
	hub        *Hub
	conn       *websocket.Conn
	send       chan Message
	ID         string
	UserID     string
	rooms      map[string]bool
	voiceGuard VoiceGuard
	buffer     AudioBuffer
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

func NewClientWithGuard(hub *Hub, conn *websocket.Conn, userID string, guard VoiceGuard, buffer AudioBuffer) *Client {
	return &Client{
		hub:        hub,
		conn:       conn,
		send:       make(chan Message, 256),
		ID:         generateClientID(),
		UserID:     userID,
		rooms:      make(map[string]bool),
		voiceGuard: guard,
		buffer:     buffer,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.releaseVoice()
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
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		if c.voiceGuard != nil && payload.IsPTT {
			if err := c.voiceGuard.StartTransmission(c.UserID, payload.Room); err != nil {
				log.Printf("PTT start rejected for user %s in room %s: %v", c.UserID, payload.Room, err)
				return
			}
		} else if c.voiceGuard != nil && !payload.IsPTT {
			c.voiceGuard.StopTransmission(c.UserID, payload.Room)
		}

		// RF06: keep the last 30s of audio in memory for replay. Buffer
		// failures never block the live relay.
		var sequence int64
		var timestamp time.Time
		if c.buffer != nil && payload.IsPTT && len(payload.Audio) > 0 {
			if audioBytes, err := decodeBase64Audio(payload.Audio); err == nil && len(audioBytes) > 0 {
				seq, ts, err := c.buffer.AddMessage(payload.Room, c.UserID, audioBytes)
				if err != nil {
					log.Printf("Audio buffer error in room %s: %v", payload.Room, err)
				} else {
					sequence, timestamp = seq, ts
				}
			}
		}

		out := map[string]interface{}{
			"audio":   payload.Audio,
			"is_ptt":  payload.IsPTT,
			"user_id": c.UserID,
			"sender":  c.ID,
		}
		if sequence > 0 {
			out["sequence"] = sequence
			out["timestamp"] = timestamp
		}
		audioPayload, _ := json.Marshal(out)
		c.hub.BroadcastToRoom(payload.Room, Message{
			Type:    "audio",
			Payload: audioPayload,
			Sender:  c,
		})

	case "request_replay":
		var payload struct {
			Room          string `json:"room"`
			SinceSequence int64  `json:"since_sequence"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		c.sendReplay(payload.Room, payload.SinceSequence)

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// releaseVoice frees the PTT lock if this client was the active speaker in any
// of its rooms, so an abrupt disconnect no longer leaves the channel busy.
func (c *Client) releaseVoice() {
	if c.voiceGuard == nil {
		return
	}

	c.hub.mu.RLock()
	rooms := make([]string, 0, len(c.rooms))
	for room := range c.rooms {
		rooms = append(rooms, room)
	}
	c.hub.mu.RUnlock()

	for _, room := range rooms {
		if err := c.voiceGuard.StopTransmission(c.UserID, room); err == nil {
			log.Printf("Released PTT lock for user %s in room %s on disconnect", c.UserID, room)
		}
	}
}

// sendReplay pushes buffered audio newer than sinceSequence to this client
// only (per-user replay via sequence numbers), followed by a terminator
// packet that closes the segment on the frontend.
func (c *Client) sendReplay(room string, sinceSequence int64) {
	if c.buffer == nil {
		return
	}

	c.hub.mu.RLock()
	_, inRoom := c.rooms[room]
	c.hub.mu.RUnlock()
	if !inRoom {
		return
	}

	msgs := c.buffer.GetMessages(room, sinceSequence)
	for _, m := range msgs {
		p, err := json.Marshal(map[string]interface{}{
			"audio":       base64.StdEncoding.EncodeToString(m.Audio),
			"is_ptt":      true,
			"user_id":     m.UserID,
			"sender":      "replay",
			"sequence":    m.Sequence,
			"timestamp":   m.Timestamp,
			"is_replayed": true,
		})
		if err != nil {
			continue
		}
		if !c.safeSend(Message{Type: "audio", Payload: p}) {
			log.Printf("Replay truncated for client %s in room %s: send buffer full", c.ID, room)
			return
		}
	}

	if len(msgs) == 0 {
		return
	}
	term, _ := json.Marshal(map[string]interface{}{
		"audio":       "",
		"is_ptt":      false,
		"sender":      "replay",
		"is_replayed": true,
	})
	if !c.safeSend(Message{Type: "audio", Payload: term}) {
		log.Printf("Replay terminator dropped for client %s in room %s", c.ID, room)
	}
}

// safeSend sends without blocking and turns "send on closed channel" (the hub
// dropped a slow consumer mid-replay) into false instead of a panic.
func (c *Client) safeSend(m Message) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	select {
	case c.send <- m:
		return true
	default:
		return false
	}
}

// decodeBase64Audio extracts the base64 string from the raw JSON audio field.
func decodeBase64Audio(raw json.RawMessage) ([]byte, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(s)
}

func generateClientID() string {
	return time.Now().Format("20060102150405.000000000")
}
