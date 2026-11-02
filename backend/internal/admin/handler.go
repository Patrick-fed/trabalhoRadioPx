// Package admin serves the RF09 administrative API: system metrics and user
// management (list/suspend/reactivate/promote/delete). Routes are gated by
// middleware.RequireAdmin; audio is never exposed or persisted (RNF01).
package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/radiopx/backend/internal/user"
	"github.com/radiopx/backend/internal/voice"
	ws "github.com/radiopx/backend/pkg/websocket"
)

// ChannelCounter is the minimal channel dependency (stats only).
type ChannelCounter interface {
	Count() (int, error)
}

// UserAdminService is the user-management boundary used by the admin handlers.
type UserAdminService interface {
	ListUsers() ([]*user.User, error)
	GetUser(userID string) (*user.User, error)
	SetUserActive(actorID, targetID string, active bool) error
	SetUserAdmin(actorID, targetID string, admin bool) error
	DeleteUser(userID string) error
}

// StatsSource gathers live dashboard counters.
type StatsSource struct {
	Users     UserAdminService
	Channels  ChannelCounter
	Hub       *ws.Hub
	Voice     *voice.VoiceService
}

type Handler struct {
	users  UserAdminService
	stats  StatsSource
}

func NewHandler(users UserAdminService, stats StatsSource) *Handler {
	return &Handler{users: users, stats: stats}
}

// IsAdmin reports whether userID holds the admin flag (used by RequireAdmin).
func (h *Handler) IsAdmin(userID string) bool {
	u, err := h.users.GetUser(userID)
	if err != nil {
		return false
	}
	return u.IsAdmin
}

type statsResponse struct {
	UsersTotal           int `json:"users_total"`
	UsersActive          int `json:"users_active"`
	ChannelsTotal        int `json:"channels_total"`
	WSConnections        int `json:"ws_connections"`
	LiveRooms            int `json:"live_rooms"`
	TransmittingChannels int `json:"transmitting_channels"`
}

// HandleStats GET /api/v1/admin/stats
func (h *Handler) HandleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	users, err := h.stats.Users.ListUsers()
	if err != nil {
		internalError(w, err)
		return
	}
	active := 0
	for _, u := range users {
		if u.IsActive {
			active++
		}
	}

	channelsTotal := 0
	if h.stats.Channels != nil {
		if n, err := h.stats.Channels.Count(); err == nil {
			channelsTotal = n
		}
	}

	wsConnections, liveRooms := 0, 0
	if h.stats.Hub != nil {
		wsConnections = h.stats.Hub.GetClientCount()
		liveRooms = h.stats.Hub.GetLiveRoomCount()
	}
	transmitting := 0
	if h.stats.Voice != nil {
		transmitting = h.stats.Voice.ActiveChannelsCount()
	}

	writeJSON(w, http.StatusOK, statsResponse{
		UsersTotal:           len(users),
		UsersActive:          active,
		ChannelsTotal:        channelsTotal,
		WSConnections:        wsConnections,
		LiveRooms:            liveRooms,
		TransmittingChannels: transmitting,
	})
}

// HandleUsers GET /api/v1/admin/users
// Password hashes are never serialized (User.Password has json:"-").
func (h *Handler) HandleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	users, err := h.users.ListUsers()
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

// HandleUser PATCH /api/v1/admin/users/{id}
// Body: {"is_active": bool} and/or {"is_admin": bool}. Self-changes are
// rejected by the service (ErrAdminSelfAction).
func (h *Handler) HandleUser(w http.ResponseWriter, r *http.Request, targetID string) {
	if r.Method != http.MethodPatch {
		methodNotAllowed(w)
		return
	}

	var req struct {
		IsActive *bool `json:"is_active"`
		IsAdmin  *bool `json:"is_admin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Corpo da requisição inválido"}`, http.StatusBadRequest)
		return
	}
	if req.IsActive == nil && req.IsAdmin == nil {
		http.Error(w, `{"error":"Informe is_active ou is_admin"}`, http.StatusBadRequest)
		return
	}

	actorID := r.Header.Get("X-User-ID")

	if req.IsActive != nil {
		if err := h.users.SetUserActive(actorID, targetID, *req.IsActive); err != nil {
			writeServiceError(w, err)
			return
		}
	}
	if req.IsAdmin != nil {
		if err := h.users.SetUserAdmin(actorID, targetID, *req.IsAdmin); err != nil {
			writeServiceError(w, err)
			return
		}
	}

	updated, err := h.users.GetUser(targetID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// HandleDeleteUser DELETE /api/v1/admin/users/{id}
func (h *Handler) HandleDeleteUser(w http.ResponseWriter, r *http.Request, targetID string) {
	if r.Method != http.MethodDelete {
		methodNotAllowed(w)
		return
	}

	actorID := r.Header.Get("X-User-ID")
	if actorID == targetID {
		http.Error(w, `{"error":"`+user.ErrAdminSelfAction.Error()+`"}`, http.StatusForbidden)
		return
	}

	if err := h.users.DeleteUser(targetID); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "User deleted"})
}

// ServeHTTP dispatches /api/v1/admin/* paths:
//
//	GET    /api/v1/admin/stats
//	GET    /api/v1/admin/users
//	PATCH  /api/v1/admin/users/{id}
//	DELETE /api/v1/admin/users/{id}
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin")
	path = strings.Trim(path, "/")

	switch {
	case path == "stats":
		h.HandleStats(w, r)
	case path == "users":
		h.HandleUsers(w, r)
	case strings.HasPrefix(path, "users/"):
		id := strings.TrimPrefix(path, "users/")
		if id == "" || strings.Contains(id, "/") {
			http.Error(w, `{"error":"User ID required"}`, http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodPatch:
			h.HandleUser(w, r, id)
		case http.MethodDelete:
			h.HandleDeleteUser(w, r, id)
		default:
			methodNotAllowed(w)
		}
	default:
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func methodNotAllowed(w http.ResponseWriter) {
	http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
}

func internalError(w http.ResponseWriter, err error) {
	http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, user.ErrUserNotFound):
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
	case errors.Is(err, user.ErrAdminSelfAction):
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusForbidden)
	default:
		internalError(w, err)
	}
}
