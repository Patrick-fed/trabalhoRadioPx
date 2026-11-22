package location

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// LocationUpdater records a user's last-known position. Implementations must
// be best-effort: errors are logged, never surfaced to the client.
type LocationUpdater interface {
	UpdateLastLocation(userID string, lat, lng float64, at time.Time) error
}

// LocationHandler exposes the dedicated periodic location endpoint (RF04):
// the app reports its GPS position every ~30s so the admin panel stays fresh
// even when the channels screen is not open (DEC-020).
type LocationHandler struct {
	locations LocationUpdater
}

func NewLocationHandler(locations LocationUpdater) *LocationHandler {
	return &LocationHandler{locations: locations}
}

// HandleUpdate accepts POST /api/v1/location with {"latitude","longitude"}
// and overwrites the user's last-known position. Best-effort: the position is
// persisted when possible, but a storage failure still returns 200.
func (h *LocationHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, `{"error":"User ID required"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		http.Error(w, `{"error":"Coordinates out of range"}`, http.StatusBadRequest)
		return
	}

	if h.locations != nil {
		if err := h.locations.UpdateLastLocation(userID, req.Latitude, req.Longitude, time.Now()); err != nil {
			log.Printf("location update user=%s: %v", userID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
