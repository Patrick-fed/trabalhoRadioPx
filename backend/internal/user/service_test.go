package user

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUserService_CreateUser(t *testing.T) {
	service := NewUserService(NewMemoryUserStore())

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
	service := NewUserService(NewMemoryUserStore())

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
	service := NewUserService(NewMemoryUserStore())

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
	service := NewUserService(NewMemoryUserStore())

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
	service := NewUserService(NewMemoryUserStore())

	_, err := service.Authenticate("nonexistent@example.com", "password123")
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	service := NewUserService(NewMemoryUserStore())

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
	service := NewUserService(NewMemoryUserStore())

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
	service := NewUserService(NewMemoryUserStore())

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
	service := NewUserService(NewMemoryUserStore())

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
	service := NewUserService(NewMemoryUserStore())

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
	service := NewUserService(NewMemoryUserStore())

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

func TestUserService_Authenticate_InactiveUser(t *testing.T) {
	service := NewUserService(NewMemoryUserStore())

	created, err := service.CreateUser(&CreateUserRequest{
		Email:    "inactive@example.com",
		Password: "password123",
		Name:     "Inactive",
	})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := service.SetUserActive("someone-else", created.ID, false); err != nil {
		t.Fatalf("SetUserActive failed: %v", err)
	}

	if _, err := service.Authenticate("inactive@example.com", "password123"); err != ErrUserInactive {
		t.Errorf("Expected ErrUserInactive, got %v", err)
	}
}

func TestUserService_SetUserActive(t *testing.T) {
	service := NewUserService(NewMemoryUserStore())

	created, _ := service.CreateUser(&CreateUserRequest{
		Email:    "target@example.com",
		Password: "password123",
		Name:     "Target",
	})
	admin, _ := service.CreateUser(&CreateUserRequest{
		Email:    "admin@example.com",
		Password: "password123",
		Name:     "Admin",
	})

	if err := service.SetUserActive(admin.ID, created.ID, false); err != nil {
		t.Fatalf("SetUserActive(false) failed: %v", err)
	}
	target, _ := service.GetUser(created.ID)
	if target.IsActive {
		t.Error("Expected user suspended")
	}

	if err := service.SetUserActive(admin.ID, created.ID, true); err != nil {
		t.Fatalf("SetUserActive(true) failed: %v", err)
	}
	target, _ = service.GetUser(created.ID)
	if !target.IsActive {
		t.Error("Expected user reactivated")
	}
}

func TestUserService_SetUserAdmin(t *testing.T) {
	service := NewUserService(NewMemoryUserStore())

	created, _ := service.CreateUser(&CreateUserRequest{
		Email:    "target@example.com",
		Password: "password123",
		Name:     "Target",
	})
	admin, _ := service.CreateUser(&CreateUserRequest{
		Email:    "admin@example.com",
		Password: "password123",
		Name:     "Admin",
	})

	if err := service.SetUserAdmin(admin.ID, created.ID, true); err != nil {
		t.Fatalf("SetUserAdmin(true) failed: %v", err)
	}
	target, _ := service.GetUser(created.ID)
	if !target.IsAdmin {
		t.Error("Expected user promoted to admin")
	}

	if err := service.SetUserAdmin(admin.ID, created.ID, false); err != nil {
		t.Fatalf("SetUserAdmin(false) failed: %v", err)
	}
	target, _ = service.GetUser(created.ID)
	if target.IsAdmin {
		t.Error("Expected admin demoted")
	}
}

func TestUserService_SetUserActive_SelfBlocked(t *testing.T) {
	service := NewUserService(NewMemoryUserStore())

	admin, _ := service.CreateUser(&CreateUserRequest{
		Email:    "admin@example.com",
		Password: "password123",
		Name:     "Admin",
	})

	if err := service.SetUserActive(admin.ID, admin.ID, false); err != ErrAdminSelfAction {
		t.Errorf("Expected ErrAdminSelfAction, got %v", err)
	}
	target, _ := service.GetUser(admin.ID)
	if !target.IsActive {
		t.Error("Admin must remain active after blocked self-suspend")
	}
}

func TestUserService_SetUserAdmin_SelfBlocked(t *testing.T) {
	service := NewUserService(NewMemoryUserStore())

	admin, _ := service.CreateUser(&CreateUserRequest{
		Email:    "admin@example.com",
		Password: "password123",
		Name:     "Admin",
	})
	admin.IsAdmin = true

	if err := service.SetUserAdmin(admin.ID, admin.ID, false); err != ErrAdminSelfAction {
		t.Errorf("Expected ErrAdminSelfAction, got %v", err)
	}
	target, _ := service.GetUser(admin.ID)
	if !target.IsAdmin {
		t.Error("Admin must keep the flag after blocked self-demote")
	}
}

func TestUserService_BootstrapAdmins(t *testing.T) {
	service := NewUserService(NewMemoryUserStore())

	existing, _ := service.CreateUser(&CreateUserRequest{
		Email:    "root@example.com",
		Password: "password123",
		Name:     "Root",
	})
	_, _ = service.CreateUser(&CreateUserRequest{
		Email:    "other@example.com",
		Password: "password123",
		Name:     "Other",
	})

	promoted, err := service.BootstrapAdmins([]string{
		" Root@Example.com ", // normaliza + trim
		"missing@example.com", // inexistente: ignorado
		"",                    // vazio: ignorado
	})
	if err != nil {
		t.Fatalf("BootstrapAdmins failed: %v", err)
	}
	if len(promoted) != 1 || promoted[0] != "root@example.com" {
		t.Errorf("Expected [root@example.com], got %v", promoted)
	}

	root, _ := service.GetUser(existing.ID)
	if !root.IsAdmin {
		t.Error("Expected root promoted to admin")
	}

	// Idempotente: segunda execução não promove de novo.
	promoted, err = service.BootstrapAdmins([]string{"root@example.com"})
	if err != nil {
		t.Fatalf("BootstrapAdmins second run failed: %v", err)
	}
	if len(promoted) != 0 {
		t.Errorf("Expected no promotions on second run, got %v", promoted)
	}
}

func TestUserService_UpdateLastLocation(t *testing.T) {
	service := NewUserService(NewMemoryUserStore())

	created, _ := service.CreateUser(&CreateUserRequest{
		Email:    "loc@example.com",
		Password: "password123",
		Name:     "Loc",
	})

	// Antes da primeira captura: sem posição.
	before, _ := service.GetUser(created.ID)
	if before.LastLat != nil || before.LastLng != nil || before.LastLocationAt != nil {
		t.Error("Expected no location before first capture")
	}

	at := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
	if err := service.UpdateLastLocation(created.ID, -23.55, -46.63, at); err != nil {
		t.Fatalf("UpdateLastLocation failed: %v", err)
	}

	after, _ := service.GetUser(created.ID)
	if after.LastLat == nil || *after.LastLat != -23.55 {
		t.Errorf("Expected last_lat -23.55, got %v", after.LastLat)
	}
	if after.LastLng == nil || *after.LastLng != -46.63 {
		t.Errorf("Expected last_lng -46.63, got %v", after.LastLng)
	}
	if after.LastLocationAt == nil || !after.LastLocationAt.Equal(at) {
		t.Errorf("Expected last_location_at %v, got %v", at, after.LastLocationAt)
	}

	// Sobrescreve (não acumula histórico).
	next := at.Add(time.Minute)
	if err := service.UpdateLastLocation(created.ID, -22.90, -43.17, next); err != nil {
		t.Fatalf("UpdateLastLocation overwrite failed: %v", err)
	}
	again, _ := service.GetUser(created.ID)
	if *again.LastLat != -22.90 || *again.LastLng != -43.17 {
		t.Errorf("Expected overwritten location, got %v/%v", again.LastLat, again.LastLng)
	}
	if !again.LastLocationAt.Equal(next) {
		t.Errorf("Expected timestamp overwritten to %v", next)
	}
}

func TestUserService_UpdateLastLocation_UnknownUser(t *testing.T) {
	service := NewUserService(NewMemoryUserStore())

	err := service.UpdateLastLocation("no-such-id", 1.0, 2.0, time.Now())
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}
