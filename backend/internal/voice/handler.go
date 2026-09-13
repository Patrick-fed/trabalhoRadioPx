package voice

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	ws "github.com/radiopx/backend/pkg/websocket"
)

type VoiceHandler struct {
	voiceService *VoiceService
	hub          *ws.Hub
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewVoiceHandler(voiceService *VoiceService, hub *ws.Hub) *VoiceHandler {
	return &VoiceHandler{
		voiceService: voiceService,
		hub:          hub,
	}
}

func (vh *VoiceHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channel")
	if channelID == "" {
		http.Error(w, `{"error":"Channel ID required"}`, http.StatusBadRequest)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "anonymous"
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := ws.NewClient(vh.hub, conn, userID)
	vh.hub.RegisterClient(client)
	vh.hub.JoinRoom(client, channelID)

	go client.WritePump()
	go client.ReadPump()

	log.Printf("Voice client connected: %s to channel %s", userID, channelID)
}

func (vh *VoiceHandler) HandleStartTransmission(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID    string `json:"user_id"`
		ChannelID string `json:"channel_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := vh.voiceService.StartTransmission(req.UserID, req.ChannelID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(VoiceStatus{
		IsTransmitting: true,
		CurrentSpeaker: req.UserID,
		ChannelID:      req.ChannelID,
	})
}

func (vh *VoiceHandler) HandleStopTransmission(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID    string `json:"user_id"`
		ChannelID string `json:"channel_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := vh.voiceService.StopTransmission(req.UserID, req.ChannelID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(VoiceStatus{
		IsTransmitting: false,
		ChannelID:      req.ChannelID,
	})
}

func (vh *VoiceHandler) HandleGetStatus(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channel")
	if channelID == "" {
		http.Error(w, `{"error":"Channel ID required"}`, http.StatusBadRequest)
		return
	}

	status := vh.voiceService.GetChannelStatus(channelID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

func (vh *VoiceHandler) BroadcastAudio(channelID string, packet *AudioPacket) {
	payload, _ := json.Marshal(packet.ToPayload())
	vh.hub.BroadcastToRoom(channelID, ws.Message{
		Type:    "audio",
		Payload: payload,
	})
}

func StartCleanupRoutine(vs *VoiceService) {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			vs.CleanupInactiveChannels(10 * time.Minute)
		}
	}()
}
