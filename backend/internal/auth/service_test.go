package auth

import (
	"testing"
)

type mockUserService struct {
	users map[string]*User
}

func newMockUserService() *mockUserService {
	return &mockUserService{
		users: make(map[string]*User),
	}
}

func (m *mockUserService) Authenticate(email, password string) (*User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, ErrInvalidCredentials
}

func (m *mockUserService) GetUser(userID string) (*User, error) {
	user, exists := m.users[userID]
	if !exists {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (m *mockUserService) CreateOAuthUser(email, name, provider, providerID string) (*User, error) {
	user := &User{
		ID:    "user1",
		Email: email,
		Name:  name,
	}
	m.users["user1"] = user
	return user, nil
}

func TestAuthService_Login(t *testing.T) {
	mockService := newMockUserService()
	mockService.users["user1"] = &User{
		ID:    "user1",
		Email: "test@example.com",
		Name:  "Test User",
	}

	authService := NewAuthService("test-secret-key-that-is-long-enough-for-jwt", mockService)

	tokenPair, err := authService.Login("test@example.com", "password")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if tokenPair.AccessToken == "" {
		t.Error("Access token should not be empty")
	}
	if tokenPair.RefreshToken == "" {
		t.Error("Refresh token should not be empty")
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	mockService := newMockUserService()
	mockService.users["user1"] = &User{
		ID:    "user1",
		Email: "test@example.com",
		Name:  "Test User",
	}

	authService := NewAuthService("test-secret-key-that-is-long-enough-for-jwt", mockService)

	tokenPair, _ := authService.Login("test@example.com", "password")

	claims, err := authService.ValidateToken(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != "user1" {
		t.Errorf("Expected user1, got %s", claims.UserID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("Expected test@example.com, got %s", claims.Email)
	}
}

func TestAuthService_ValidateToken_Invalid(t *testing.T) {
	mockService := newMockUserService()
	authService := NewAuthService("test-secret-key-that-is-long-enough-for-jwt", mockService)

	_, err := authService.ValidateToken("invalid-token")
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	mockService := newMockUserService()
	mockService.users["user1"] = &User{
		ID:    "user1",
		Email: "test@example.com",
		Name:  "Test User",
	}

	authService := NewAuthService("test-secret-key-that-is-long-enough-for-jwt", mockService)

	tokenPair, _ := authService.Login("test@example.com", "password")

	newTokenPair, err := authService.RefreshToken(tokenPair.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}

	if newTokenPair.AccessToken == "" {
		t.Error("New access token should not be empty")
	}
}

func TestAuthService_GetUserIDFromToken(t *testing.T) {
	mockService := newMockUserService()
	mockService.users["user1"] = &User{
		ID:    "user1",
		Email: "test@example.com",
		Name:  "Test User",
	}

	authService := NewAuthService("test-secret-key-that-is-long-enough-for-jwt", mockService)

	tokenPair, _ := authService.Login("test@example.com", "password")

	userID, err := authService.GetUserIDFromToken(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("GetUserIDFromToken failed: %v", err)
	}

	if userID != "user1" {
		t.Errorf("Expected user1, got %s", userID)
	}
}
