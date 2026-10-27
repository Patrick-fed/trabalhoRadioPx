package db_test

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/radiopx/backend/internal/channel"
	"github.com/radiopx/backend/internal/db"
	"github.com/radiopx/backend/internal/user"
)

// TestPostgresStores exercises the real PG stores and the migration runner
// against the database configured by DB_* (backend/.env). It is skipped unless
// RADIOPX_TEST_DB=1 so the default `go test ./...` stays database-free.
func TestPostgresStores(t *testing.T) {
	if os.Getenv("RADIOPX_TEST_DB") == "" {
		t.Skip("set RADIOPX_TEST_DB=1 to run PostgreSQL integration tests")
	}
	loadDotEnv(t)

	database, err := db.Open(db.ConfigFromEnv())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer database.Close()

	if err := db.ApplyMigrations(database); err != nil {
		t.Fatalf("first migration run: %v", err)
	}
	if err := db.ApplyMigrations(database); err != nil {
		t.Fatalf("second migration run (must be idempotent): %v", err)
	}

	userStore := user.NewPGUserStore(database)
	channelStore := channel.NewPGChannelStore(database)
	userService := user.NewUserService(userStore)
	channelService := channel.NewChannelService(channelStore)

	createdUsers := make([]string, 0)
	createdChannels := make([]string, 0)
	defer func() {
		for _, id := range createdChannels {
			channelStore.Delete(id)
		}
		for _, id := range createdUsers {
			userStore.Delete(id)
		}
	}()

	newUser := func(email, name string) *user.User {
		t.Helper()
		u, err := userService.CreateUser(&user.CreateUserRequest{
			Email:    email,
			Password: "password123",
			Name:     name,
		})
		if err != nil {
			t.Fatalf("CreateUser(%s): %v", email, err)
		}
		createdUsers = append(createdUsers, u.ID)
		return u
	}

	// Unique per run so leftover rows from a previous (interrupted) run
	// cannot collide on the email UNIQUE constraint.
	runID := time.Now().Format("20060102150405.000000")
	email := func(local string) string { return local + "-" + runID + "@example.com" }

	// --- user store ---
	owner := newUser(email("it-owner"), "Owner")

	if _, err := userService.GetUserByEmail(strings.ToUpper(email("it-owner"))); err != nil {
		t.Errorf("email must be normalized before lookup: %v", err)
	}
	if _, err := userService.CreateUser(&user.CreateUserRequest{
		Email: email("it-owner"), Password: "password123", Name: "Dup",
	}); err != user.ErrUserAlreadyExists {
		t.Errorf("duplicate email: got %v, want ErrUserAlreadyExists", err)
	}
	if _, err := userService.Authenticate(email("it-owner"), "wrongpass1"); err != user.ErrInvalidCredentials {
		t.Errorf("bad password: got %v, want ErrInvalidCredentials", err)
	}
	if _, err := userService.Authenticate(email("it-owner"), "password123"); err != nil {
		t.Errorf("valid login: %v", err)
	}

	// --- channel store: membership round-trip ---
	ch, err := channelService.CreateChannel(owner.ID, &channel.CreateChannelRequest{
		Name: "IT Channel", Latitude: -23.5505, Longitude: -46.6333,
	})
	if err != nil {
		t.Fatalf("CreateChannel: %v", err)
	}
	createdChannels = append(createdChannels, ch.ID)

	got, err := channelService.GetChannel(ch.ID)
	if err != nil {
		t.Fatalf("GetChannel: %v", err)
	}
	if got.UserCount != 1 || !got.IsAvailable {
		t.Errorf("fresh channel: user_count=%d is_available=%v, want 1/true", got.UserCount, got.IsAvailable)
	}

	guest := newUser(email("it-guest"), "Guest")
	if err := channelService.JoinChannel(guest.ID, ch.ID); err != nil {
		t.Fatalf("JoinChannel: %v", err)
	}
	if err := channelService.JoinChannel(guest.ID, ch.ID); err != channel.ErrUserAlreadyInChannel {
		t.Errorf("duplicate join: got %v, want ErrUserAlreadyInChannel", err)
	}
	if err := channelService.LeaveChannel(guest.ID, ch.ID); err != nil {
		t.Fatalf("LeaveChannel: %v", err)
	}
	if err := channelService.LeaveChannel(guest.ID, ch.ID); err != channel.ErrUserNotInChannel {
		t.Errorf("double leave: got %v, want ErrUserNotInChannel", err)
	}

	// --- channel store: capacity enforced transactionally ---
	full, err := channelService.CreateChannel(owner.ID, &channel.CreateChannelRequest{
		Name: "IT Full Channel", Latitude: -23.56, Longitude: -46.64,
	})
	if err != nil {
		t.Fatalf("CreateChannel(full): %v", err)
	}
	createdChannels = append(createdChannels, full.ID)

	for i := 0; i < channel.MaxUsersPerChannel-1; i++ {
		member := newUser(email(fmt.Sprintf("it-member-%d", i)), "Member")
		if err := channelService.JoinChannel(member.ID, full.ID); err != nil {
			t.Fatalf("join %d: %v", i, err)
		}
	}
	extra := newUser(email("it-extra"), "Extra")
	if err := channelService.JoinChannel(extra.ID, full.ID); err != channel.ErrChannelFull {
		t.Errorf("join full channel: got %v, want ErrChannelFull", err)
	}

	// --- channel store: queries ---
	nearby, err := channelService.GetNearbyChannels(-23.5505, -46.6333, 100.0)
	if err != nil {
		t.Fatalf("GetNearbyChannels: %v", err)
	}
	found := false
	for _, c := range nearby {
		if c.ID == ch.ID {
			found = true
		}
	}
	if !found {
		t.Errorf("GetNearbyChannels did not return created channel")
	}

	mine, err := channelService.GetUserChannels(owner.ID)
	if err != nil {
		t.Fatalf("GetUserChannels: %v", err)
	}
	if len(mine) < 2 {
		t.Errorf("GetUserChannels: got %d, want at least 2 (both channels created by owner)", len(mine))
	}

	if err := channelService.UpdateChannel(ch.ID, guest.ID, &channel.UpdateChannelRequest{Name: "Hacked"}); err == nil {
		t.Error("UpdateChannel by non-creator must fail")
	}
	if err := channelService.UpdateChannel(ch.ID, owner.ID, &channel.UpdateChannelRequest{Name: "Renamed"}); err != nil {
		t.Errorf("UpdateChannel by creator: %v", err)
	}

	if _, err := channelService.GetChannel("00000000-0000-0000-0000-000000000000"); err != channel.ErrChannelNotFound {
		t.Errorf("missing channel: got %v, want ErrChannelNotFound", err)
	}
}

// loadDotEnv fills DB_* from backend/.env when the process environment does
// not provide them, mirroring the server's loadEnvFile().
func loadDotEnv(t *testing.T) {
	t.Helper()
	for _, path := range []string{"../../.env", ".env"} {
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
				continue
			}
			key, value, _ := strings.Cut(line, "=")
			key = strings.TrimSpace(key)
			if key != "" && os.Getenv(key) == "" {
				os.Setenv(key, strings.Trim(strings.TrimSpace(value), `"'`))
			}
		}
		file.Close()
		return
	}
}
