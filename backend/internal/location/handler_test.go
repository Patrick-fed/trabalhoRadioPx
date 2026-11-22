package location

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/radiopx/backend/internal/user"
)

func newTestHandler(t *testing.T) (*LocationHandler, *user.UserService, *user.User) {
	t.Helper()
	svc := user.NewUserService(user.NewMemoryUserStore())
	u, err := svc.CreateUser(&user.CreateUserRequest{
		Email:    "loc@example.com",
		Password: "password123",
		Name:     "Local",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return NewLocationHandler(svc), svc, u
}

func doReq(t *testing.T, h *LocationHandler, method, body, userID string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, "/api/v1/location", bytes.NewBufferString(body))
	if userID != "" {
		req.Header.Set("X-User-ID", userID)
	}
	rec := httptest.NewRecorder()
	h.HandleUpdate(rec, req)
	return rec
}

func TestLocationHandler_Update_OK(t *testing.T) {
	h, svc, u := newTestHandler(t)

	rec := doReq(t, h, http.MethodPost, `{"latitude":-23.55,"longitude":-46.63}`, u.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	after, err := svc.GetUser(u.ID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if after.LastLat == nil || *after.LastLat != -23.55 {
		t.Errorf("expected last_lat -23.55, got %v", after.LastLat)
	}
	if after.LastLng == nil || *after.LastLng != -46.63 {
		t.Errorf("expected last_lng -46.63, got %v", after.LastLng)
	}
	if after.LastLocationAt == nil || time.Since(*after.LastLocationAt) > time.Minute {
		t.Errorf("expected fresh last_location_at, got %v", after.LastLocationAt)
	}
}

func TestLocationHandler_Update_OverwritesPosition(t *testing.T) {
	h, svc, u := newTestHandler(t)

	doReq(t, h, http.MethodPost, `{"latitude":1.0,"longitude":2.0}`, u.ID)
	doReq(t, h, http.MethodPost, `{"latitude":3.0,"longitude":4.0}`, u.ID)

	after, err := svc.GetUser(u.ID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if *after.LastLat != 3.0 || *after.LastLng != 4.0 {
		t.Errorf("expected 3.0/4.0, got %v/%v", after.LastLat, after.LastLng)
	}
}

func TestLocationHandler_NoUserID(t *testing.T) {
	h, _, _ := newTestHandler(t)
	rec := doReq(t, h, http.MethodPost, `{"latitude":1.0,"longitude":2.0}`, "")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestLocationHandler_InvalidBody(t *testing.T) {
	h, _, u := newTestHandler(t)
	rec := doReq(t, h, http.MethodPost, `not-json`, u.ID)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestLocationHandler_CoordinatesOutOfRange(t *testing.T) {
	h, _, u := newTestHandler(t)
	cases := []string{
		`{"latitude":91,"longitude":0}`,
		`{"latitude":-91,"longitude":0}`,
		`{"latitude":0,"longitude":181}`,
		`{"latitude":0,"longitude":-181}`,
	}
	for _, body := range cases {
		rec := doReq(t, h, http.MethodPost, body, u.ID)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("body %s: expected 400, got %d", body, rec.Code)
		}
	}
}

func TestLocationHandler_MethodNotAllowed(t *testing.T) {
	h, _, u := newTestHandler(t)
	rec := doReq(t, h, http.MethodGet, "", u.ID)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
