package user

import (
	"testing"

	"github.com/google/uuid"
)

func TestUserService_CreateUser(t *testing.T) {
	service := NewUserService()

	req := &CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	user, err := service.CreateUser(req)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, user.Email)
	}
	if user.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, user.Name)
	}
	// Password should be hashed, not exposed
	if user.Password == req.Password {
		t.Error("Password should be hashed, not exposed")
	}
}

func TestUserService_CreateUser_DuplicateEmail(t *testing.T) {
	service := NewUserService()

	req := &CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	service.CreateUser(req)

	_, err := service.CreateUser(req)
	if err != ErrUserAlreadyExists {
		t.Errorf("Expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestUserService_Authenticate(t *testing.T) {
	service := NewUserService()

	req := &CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	service.CreateUser(req)

	user, err := service.Authenticate("test@example.com", "password123")
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}

	if user.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, user.Email)
	}
}

func TestUserService_Authenticate_WrongPassword(t *testing.T) {
	service := NewUserService()

	req := &CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	service.CreateUser(req)

	_, err := service.Authenticate("test@example.com", "wrongpassword")
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserService_Authenticate_UserNotFound(t *testing.T) {
	service := NewUserService()

	_, err := service.Authenticate("nonexistent@example.com", "password123")
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	service := NewUserService()

	req := &CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	user, _ := service.CreateUser(req)

	err := service.UpdateUser(user.ID, &UpdateUserRequest{
		Name: "Updated Name",
	})

	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	updated, _ := service.GetUser(user.ID)
	if updated.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got %s", updated.Name)
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	service := NewUserService()

	req := &CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	user, _ := service.CreateUser(req)

	err := service.DeleteUser(user.ID)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	_, err = service.GetUser(user.ID)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_ChangePassword(t *testing.T) {
	service := NewUserService()

	req := &CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	user, _ := service.CreateUser(req)

	err := service.ChangePassword(user.ID, "password123", "newpassword123")
	if err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}

	_, err = service.Authenticate("test@example.com", "password123")
	if err != ErrInvalidCredentials {
		t.Errorf("Old password should not work anymore")
	}

	_, err = service.Authenticate("test@example.com", "newpassword123")
	if err != nil {
		t.Errorf("New password should work")
	}
}

func TestUserService_ChangePassword_TooShort(t *testing.T) {
	service := NewUserService()

	user, _ := service.CreateUser(&CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	})

	err := service.ChangePassword(user.ID, "password123", "short")
	if err != ErrPasswordTooShort {
		t.Errorf("Expected ErrPasswordTooShort, got %v", err)
	}

	// The old password must remain valid after a rejected change.
	if _, err := service.Authenticate("test@example.com", "password123"); err != nil {
		t.Errorf("Old password should still work after rejected change, got %v", err)
	}
}

func TestUserService_ChangePassword_WrongOldPassword(t *testing.T) {
	service := NewUserService()

	user, _ := service.CreateUser(&CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	})

	err := service.ChangePassword(user.ID, "wrongpassword", "newpassword123")
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserService_CreateUser_GeneratesUniqueUUIDs(t *testing.T) {
	service := NewUserService()

	first, err := service.CreateUser(&CreateUserRequest{
		Email:    "first@example.com",
		Password: "password123",
		Name:     "First",
	})
	if err != nil {
		t.Fatalf("CreateUser first failed: %v", err)
	}

	second, err := service.CreateUser(&CreateUserRequest{
		Email:    "second@example.com",
		Password: "password123",
		Name:     "Second",
	})
	if err != nil {
		t.Fatalf("CreateUser second failed: %v", err)
	}

	if _, err := uuid.Parse(first.ID); err != nil {
		t.Errorf("Expected a valid UUID, got %q: %v", first.ID, err)
	}
	if first.ID == second.ID {
		t.Errorf("Expected distinct user IDs, both were %q", first.ID)
	}
}
