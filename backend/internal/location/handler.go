package location

import (
	"encoding/json"
	"net/http"
)

type LocationHandler struct {
	locationService *LocationService
}

func NewLocationHandler(locationService *LocationService) *LocationHandler {
	return &LocationHandler{
		locationService: locationService,
	}
}

func (h *LocationHandler) HandleUpdateLocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID    string           `json:"user_id"`
		Location  *LocationUpdate  `json:"location"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		http.Error(w, `{"error":"User ID required"}`, http.StatusBadRequest)
		return
	}

	if req.Location == nil {
		http.Error(w, `{"error":"Location required"}`, http.StatusBadRequest)
		return
	}

	location, err := h.locationService.UpdateUserLocation(req.UserID, req.Location)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(location)
}

func (h *LocationHandler) HandleGetNearbyChannels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, `{"error":"User ID required"}`, http.StatusBadRequest)
		return
	}

	radiusKm := DefaultRadiusKm
	if r := r.URL.Query().Get("radius"); r != "" {
		if parsed, err := parseFloat(r); err == nil {
			radiusKm = parsed
		}
	}

	filter := &ProximityFilter{
		RadiusKm:   radiusKm,
		MaxResults: MaxVisibleChannels,
	}

	nearbyChannels, err := h.locationService.FindNearbyChannels(userID, filter)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(nearbyChannels)
}

func (h *LocationHandler) HandleGetUsersInRadius(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	latitude, err := parseFloat(r.URL.Query().Get("lat"))
	if err != nil {
		http.Error(w, `{"error":"Invalid latitude"}`, http.StatusBadRequest)
		return
	}

	longitude, err := parseFloat(r.URL.Query().Get("lng"))
	if err != nil {
		http.Error(w, `{"error":"Invalid longitude"}`, http.StatusBadRequest)
		return
	}

	radiusKm := DefaultRadiusKm
	if r := r.URL.Query().Get("radius"); r != "" {
		if parsed, err := parseFloat(r); err == nil {
			radiusKm = parsed
		}
	}

	users := h.locationService.GetUsersInRadius(latitude, longitude, radiusKm)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
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
