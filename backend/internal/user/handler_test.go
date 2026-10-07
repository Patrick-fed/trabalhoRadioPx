package user

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUserHandler_HandleUpdateUserProfile(t *testing.T) {
	service := NewUserService()
	handler := NewUserHandler(service)

	created, err := service.CreateUser(&CreateUserRequest{
		Email:    "profile@example.com",
		Password: "password123",
		Name:     "Old Name",
	})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/user/profile",
		strings.NewReader(`{"name":"New Name"}`),
	)
	rec := httptest.NewRecorder()

	handler.HandleUpdateUserProfile(rec, req, created.ID)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	// Regression guard: the frontend parses this body with ProfileModel.fromJson,
	// which needs id and email. Returning only {"message": ...} used to throw
	// "type 'Null' is not a subtype of type 'String'" in the app.
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("Response is not valid JSON: %v", err)
	}

	if id, _ := got["id"].(string); id == "" {
		t.Errorf("Expected response to include id, got %v", got["id"])
	}
	if email, _ := got["email"].(string); email != "profile@example.com" {
		t.Errorf("Expected email profile@example.com, got %v", got["email"])
	}
	if name, _ := got["name"].(string); name != "New Name" {
		t.Errorf("Expected name New Name, got %v", got["name"])
	}
	if _, exists := got["password"]; exists {
		t.Error("Response must not expose the password field")
	}

	updated, err := service.GetUser(created.ID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("Expected stored name New Name, got %s", updated.Name)
	}
}

func TestUserHandler_HandleUpdateUserProfile_NotFound(t *testing.T) {
	service := NewUserService()
	handler := NewUserHandler(service)

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/user/profile",
		strings.NewReader(`{"name":"New Name"}`),
	)
	rec := httptest.NewRecorder()

	handler.HandleUpdateUserProfile(rec, req, "missing-user-id")

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}
}
