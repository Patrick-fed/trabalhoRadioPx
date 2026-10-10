package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/radiopx/backend/internal/auth"
	"github.com/radiopx/backend/internal/user"
)

const testSecret = "test-secret-key-that-is-long-enough-for-jwt"

func newTestAuthEnv() (*user.UserService, *auth.AuthService) {
	userService := user.NewUserService()
	authAdapter := &authUserServiceAdapter{svc: userService}
	authService := auth.NewAuthService(testSecret, authAdapter)
	return userService, authService
}

func doRegister(t *testing.T, userService *user.UserService, authService *auth.AuthService, payload string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(payload))
	rec := httptest.NewRecorder()
	handleRegister(rec, req, userService, authService)
	return rec
}

func doLogin(t *testing.T, authService *auth.AuthService, payload string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(payload))
	rec := httptest.NewRecorder()
	handleLogin(rec, req, authService)
	return rec
}

func registerAndDecode(t *testing.T, userService *user.UserService, authService *auth.AuthService, email, password, name string) map[string]any {
	t.Helper()
	payload := `{"email":` + marshalString(email) + `,"password":` + marshalString(password) + `,"name":` + marshalString(name) + `}`
	rec := doRegister(t, userService, authService, payload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected register 201, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("Register response is not valid JSON: %v", err)
	}
	return body
}

func marshalString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func TestHandleRegister_Success(t *testing.T) {
	userService, authService := newTestAuthEnv()

	body := registerAndDecode(t, userService, authService, "register@example.com", "password123", "New User")

	if token, _ := body["token"].(string); token == "" {
		t.Error("Expected access token in register response")
	}
	if refresh, _ := body["refresh_token"].(string); refresh == "" {
		t.Error("Expected refresh token in register response")
	}
	userMap, ok := body["user"].(map[string]any)
	if !ok {
		t.Fatal("Expected user object in register response")
	}
	if email, _ := userMap["email"].(string); email != "register@example.com" {
		t.Errorf("Expected email register@example.com, got %v", userMap["email"])
	}
	if _, exists := userMap["password"]; exists {
		t.Error("Register response must not expose the password")
	}
}

func TestHandleRegister_DuplicateEmail(t *testing.T) {
	userService, authService := newTestAuthEnv()
	registerAndDecode(t, userService, authService, "dupe@example.com", "password123", "A")

	rec := doRegister(t, userService, authService, `{"email":"dupe@example.com","password":"password123","name":"B"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("Expected 409, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("já está cadastrado")) {
		t.Errorf("Expected PT-BR duplicate message, got %s", rec.Body.String())
	}
}

func TestHandleRegister_InvalidInputs(t *testing.T) {
	userService, authService := newTestAuthEnv()

	cases := []struct {
		name    string
		payload string
		want    int
	}{
		{"missing email", `{"password":"password123","name":"A"}`, http.StatusBadRequest},
		{"missing password", `{"email":"a@example.com","name":"A"}`, http.StatusBadRequest},
		{"short password", `{"email":"short@example.com","password":"12345","name":"A"}`, http.StatusBadRequest},
		{"invalid email", `{"email":"not-an-email","password":"password123","name":"A"}`, http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doRegister(t, userService, authService, tc.payload)
			if rec.Code != tc.want {
				t.Errorf("Expected %d, got %d (body: %s)", tc.want, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandleLogin_Success(t *testing.T) {
	userService, authService := newTestAuthEnv()
	registerAndDecode(t, userService, authService, "login@example.com", "password123", "Login User")

	rec := doLogin(t, authService, `{"email":"login@example.com","password":"password123"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("Login response is not valid JSON: %v", err)
	}
	if token, _ := body["token"].(string); token == "" {
		t.Error("Expected access token in login response")
	}
}

func TestHandleLogin_WrongPassword(t *testing.T) {
	userService, authService := newTestAuthEnv()
	registerAndDecode(t, userService, authService, "wrong@example.com", "password123", "A")

	rec := doLogin(t, authService, `{"email":"wrong@example.com","password":"wrongpass"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("E-mail ou senha inválidos")) {
		t.Errorf("Expected PT-BR invalid credentials message, got %s", rec.Body.String())
	}
}

func TestHandleLogin_EmailNormalized(t *testing.T) {
	userService, authService := newTestAuthEnv()
	registerAndDecode(t, userService, authService, "Mixed@Example.COM", "password123", "Mixed Case")

	rec := doLogin(t, authService, `{"email":"  mixed@example.com  ","password":"password123"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 after normalization, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestHandleRefresh_RejectsAccessToken(t *testing.T) {
	userService, authService := newTestAuthEnv()
	body := registerAndDecode(t, userService, authService, "refresh@example.com", "password123", "A")
	accessToken, _ := body["token"].(string)
	if accessToken == "" {
		t.Fatal("Expected access token")
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBufferString(`{"refresh_token":`+marshalString(accessToken)+`}`))
	rec := httptest.NewRecorder()
	handleRefresh(rec, req, authService)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401 for access token used as refresh, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestHandleRefresh_WithRefreshToken(t *testing.T) {
	userService, authService := newTestAuthEnv()
	body := registerAndDecode(t, userService, authService, "refresh2@example.com", "password123", "A")
	refreshToken, _ := body["refresh_token"].(string)
	if refreshToken == "" {
		t.Fatal("Expected refresh token")
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBufferString(`{"refresh_token":`+marshalString(refreshToken)+`}`))
	rec := httptest.NewRecorder()
	handleRefresh(rec, req, authService)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}
