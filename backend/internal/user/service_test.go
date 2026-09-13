package user

import (
	"testing"
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
