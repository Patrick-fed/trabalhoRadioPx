package channel

import (
	"encoding/json"
	"net/http"
	"strings"
)

type ChannelHandler struct {
	channelService *ChannelService
}

func NewChannelHandler(channelService *ChannelService) *ChannelHandler {
	return &ChannelHandler{
		channelService: channelService,
	}
}

func (h *ChannelHandler) HandleCreateChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req CreateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, `{"error":"User ID required"}`, http.StatusBadRequest)
		return
	}

	channel, err := h.channelService.CreateChannel(userID, &req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(channel)
}

func (h *ChannelHandler) HandleGetChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	channelID := extractChannelID(r.URL.Path)
	if channelID == "" {
		http.Error(w, `{"error":"Channel ID required"}`, http.StatusBadRequest)
		return
	}

	channel, err := h.channelService.GetChannel(channelID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(channel)
}

func (h *ChannelHandler) HandleUpdateChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	channelID := extractChannelID(r.URL.Path)
	if channelID == "" {
		http.Error(w, `{"error":"Channel ID required"}`, http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, `{"error":"User ID required"}`, http.StatusBadRequest)
		return
	}

	var req UpdateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.channelService.UpdateChannel(channelID, userID, &req); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Channel updated"})
}

func (h *ChannelHandler) HandleDeleteChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	channelID := extractChannelID(r.URL.Path)
	if channelID == "" {
		http.Error(w, `{"error":"Channel ID required"}`, http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, `{"error":"User ID required"}`, http.StatusBadRequest)
		return
	}

	if err := h.channelService.DeleteChannel(channelID, userID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Channel deleted"})
}

func (h *ChannelHandler) HandleJoinChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	channelID := extractChannelID(r.URL.Path)
	if channelID == "" {
		http.Error(w, `{"error":"Channel ID required"}`, http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, `{"error":"User ID required"}`, http.StatusBadRequest)
		return
	}

	if err := h.channelService.JoinChannel(userID, channelID); err != nil {
		if err == ErrUserAlreadyInChannel {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"message": "Already in channel"})
			return
		}
		status := http.StatusConflict
		if err == ErrChannelNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Joined channel"})
}

func (h *ChannelHandler) HandleLeaveChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	channelID := extractChannelID(r.URL.Path)
	if channelID == "" {
		http.Error(w, `{"error":"Channel ID required"}`, http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, `{"error":"User ID required"}`, http.StatusBadRequest)
		return
	}

	if err := h.channelService.LeaveChannel(userID, channelID); err != nil {
		status := http.StatusConflict
		if err == ErrChannelNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Left channel"})
}

func (h *ChannelHandler) HandleGetNearbyChannels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	lat, err := parseFloat(r.URL.Query().Get("lat"))
	if err != nil {
		http.Error(w, `{"error":"Invalid latitude"}`, http.StatusBadRequest)
		return
	}

	lng, err := parseFloat(r.URL.Query().Get("lng"))
	if err != nil {
		http.Error(w, `{"error":"Invalid longitude"}`, http.StatusBadRequest)
		return
	}

	radiusKm := 10.0
	if r := r.URL.Query().Get("radius"); r != "" {
		if parsed, err := parseFloat(r); err == nil {
			radiusKm = parsed
		}
	}

	channels, err := h.channelService.GetNearbyChannels(lat, lng, radiusKm)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(channels)
}

func (h *ChannelHandler) HandleGetUserChannels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, `{"error":"User ID required"}`, http.StatusBadRequest)
		return
	}

	channels := h.channelService.GetUserChannels(userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(channels)
}

func (h *ChannelHandler) HandleGetChannelUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	channelID := extractChannelID(r.URL.Path)
	if channelID == "" {
		http.Error(w, `{"error":"Channel ID required"}`, http.StatusBadRequest)
		return
	}

	users, err := h.channelService.GetChannelUsers(channelID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func extractChannelID(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "channels" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := parseFloatValue(s, &f)
	return f, err
}

func parseFloatValue(s string, f *float64) (int, error) {
	n := 0
	sign := 1.0

	if n < len(s) && s[n] == '-' {
		sign = -1.0
		n++
	}

	for n < len(s) && s[n] >= '0' && s[n] <= '9' {
		*f = *f*10 + float64(s[n]-'0')
		n++
	}

	if n < len(s) && s[n] == '.' {
		n++
		decimal := 0.1
		for n < len(s) && s[n] >= '0' && s[n] <= '9' {
			*f += float64(s[n]-'0') * decimal
			decimal /= 10
			n++
		}
	}

	*f *= sign

	return n, nil
}
