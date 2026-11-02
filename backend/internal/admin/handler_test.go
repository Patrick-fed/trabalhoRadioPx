package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/radiopx/backend/internal/middleware"
	"github.com/radiopx/backend/internal/user"
)

// fakeUsers implements UserAdminService over the in-memory user store.
type fakeUsers struct {
	svc *user.UserService
}

func (f *fakeUsers) ListUsers() ([]*user.User, error)     { return f.svc.ListUsers() }
func (f *fakeUsers) GetUser(id string) (*user.User, error) { return f.svc.GetUser(id) }
func (f *fakeUsers) SetUserActive(a, t string, v bool) error {
	return f.svc.SetUserActive(a, t, v)
}
func (f *fakeUsers) SetUserAdmin(a, t string, v bool) error {
	return f.svc.SetUserAdmin(a, t, v)
}
func (f *fakeUsers) DeleteUser(id string) error { return f.svc.DeleteUser(id) }

func newTestEnv(t *testing.T) (*Handler, *user.UserService, *user.User, *user.User) {
	t.Helper()
	svc := user.NewUserService(user.NewMemoryUserStore())

	adminUser, err := svc.CreateUser(&user.CreateUserRequest{
		Email:    "admin@example.com",
		Password: "password123",
		Name:     "Admin",
	})
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}
	adminUser.IsAdmin = true

	regular, err := svc.CreateUser(&user.CreateUserRequest{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "Regular",
	})
	if err != nil {
		t.Fatalf("create regular: %v", err)
	}

	fu := &fakeUsers{svc: svc}
	h := NewHandler(fu, StatsSource{Users: fu})
	return h, svc, adminUser, regular
}

// withAdmin wraps the handler with RequireAdmin using the test handler's IsAdmin.
func withAdmin(h *Handler) http.Handler {
	return middleware.RequireAdmin(h.IsAdmin)(h)
}

func doReq(t *testing.T, handler http.Handler, method, path, body, userID string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Buffer
	if body == "" {
		reader = bytes.NewBuffer(nil)
	} else {
		reader = bytes.NewBufferString(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if userID != "" {
		req.Header.Set("X-User-ID", userID)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestAdmin_RequireAdmin_NoUser(t *testing.T) {
	h, _, _, _ := newTestEnv(t)
	rec := doReq(t, withAdmin(h), http.MethodGet, "/api/v1/admin/stats", "", "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", rec.Code)
	}
}

func TestAdmin_RequireAdmin_NotAdmin(t *testing.T) {
	h, _, _, regular := newTestEnv(t)
	rec := doReq(t, withAdmin(h), http.MethodGet, "/api/v1/admin/stats", "", regular.ID)
	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected 403, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestAdmin_Stats_OK(t *testing.T) {
	h, _, adminUser, _ := newTestEnv(t)
	rec := doReq(t, withAdmin(h), http.MethodGet, "/api/v1/admin/stats", "", adminUser.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	var payload statsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if payload.UsersTotal != 2 {
		t.Errorf("Expected users_total=2, got %d", payload.UsersTotal)
	}
	if payload.UsersActive != 2 {
		t.Errorf("Expected users_active=2, got %d", payload.UsersActive)
	}
	if payload.ChannelsTotal != 0 {
		t.Errorf("Expected channels_total=0, got %d", payload.ChannelsTotal)
	}
}

func TestAdmin_ListUsers_HidesPassword(t *testing.T) {
	h, _, adminUser, _ := newTestEnv(t)
	rec := doReq(t, withAdmin(h), http.MethodGet, "/api/v1/admin/users", "", adminUser.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", rec.Code)
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("password")) {
		t.Error("Response must not contain password field")
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("password_hash")) {
		t.Error("Response must not contain password_hash")
	}

	var users []*user.User
	if err := json.Unmarshal(rec.Body.Bytes(), &users); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestAdmin_Patch_SuspendAndReactivate(t *testing.T) {
	h, _, adminUser, regular := newTestEnv(t)

	rec := doReq(t, withAdmin(h), http.MethodPatch, "/api/v1/admin/users/"+regular.ID,
		`{"is_active":false}`, adminUser.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("suspend: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	// Suspended user cannot authenticate.
	if _, err := h.users.(*fakeUsers).svc.Authenticate("user@example.com", "password123"); err != user.ErrUserInactive {
		t.Errorf("Expected ErrUserInactive after suspend, got %v", err)
	}

	rec = doReq(t, withAdmin(h), http.MethodPatch, "/api/v1/admin/users/"+regular.ID,
		`{"is_active":true}`, adminUser.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("reactivate: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if _, err := h.users.(*fakeUsers).svc.Authenticate("user@example.com", "password123"); err != nil {
		t.Errorf("Expected login after reactivate, got %v", err)
	}
}

func TestAdmin_Patch_SelfSuspendBlocked(t *testing.T) {
	h, _, adminUser, _ := newTestEnv(t)
	rec := doReq(t, withAdmin(h), http.MethodPatch, "/api/v1/admin/users/"+adminUser.ID,
		`{"is_active":false}`, adminUser.ID)
	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected 403 on self-suspend, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestAdmin_Patch_PromoteAndDemote(t *testing.T) {
	h, _, adminUser, regular := newTestEnv(t)

	rec := doReq(t, withAdmin(h), http.MethodPatch, "/api/v1/admin/users/"+regular.ID,
		`{"is_admin":true}`, adminUser.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("promote: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	target, _ := h.users.GetUser(regular.ID)
	if !target.IsAdmin {
		t.Error("Expected regular user promoted")
	}

	rec = doReq(t, withAdmin(h), http.MethodPatch, "/api/v1/admin/users/"+regular.ID,
		`{"is_admin":false}`, adminUser.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("demote: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	target, _ = h.users.GetUser(regular.ID)
	if target.IsAdmin {
		t.Error("Expected user demoted")
	}
}

func TestAdmin_Patch_EmptyBody(t *testing.T) {
	h, _, adminUser, regular := newTestEnv(t)
	rec := doReq(t, withAdmin(h), http.MethodPatch, "/api/v1/admin/users/"+regular.ID,
		`{}`, adminUser.ID)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 on empty patch, got %d", rec.Code)
	}
}

func TestAdmin_Delete_User(t *testing.T) {
	h, _, adminUser, regular := newTestEnv(t)
	rec := doReq(t, withAdmin(h), http.MethodDelete, "/api/v1/admin/users/"+regular.ID,
		"", adminUser.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if _, err := h.users.GetUser(regular.ID); err != user.ErrUserNotFound {
		t.Errorf("Expected user deleted, got %v", err)
	}
}

func TestAdmin_Delete_SelfBlocked(t *testing.T) {
	h, _, adminUser, _ := newTestEnv(t)
	rec := doReq(t, withAdmin(h), http.MethodDelete, "/api/v1/admin/users/"+adminUser.ID,
		"", adminUser.ID)
	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected 403 on self-delete, got %d", rec.Code)
	}
}

func TestAdmin_Delete_NotFound(t *testing.T) {
	h, _, adminUser, _ := newTestEnv(t)
	rec := doReq(t, withAdmin(h), http.MethodDelete, "/api/v1/admin/users/no-such-id",
		"", adminUser.ID)
	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestAdmin_Route_NotFound(t *testing.T) {
	h, _, adminUser, _ := newTestEnv(t)
	rec := doReq(t, withAdmin(h), http.MethodGet, "/api/v1/admin/nope", "", adminUser.ID)
	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}
