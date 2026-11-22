package channel

import (
	"fmt"
	"sync"
	"testing"
)

func TestChannelService_CreateChannel(t *testing.T) {
	service := NewChannelService(NewMemoryChannelStore())

	req := &CreateChannelRequest{
		Name:      "Test Channel",
		Latitude:  -23.5505,
		Longitude: -46.6333,
	}

	channel, err := service.CreateChannel("user1", req)
	if err != nil {
		t.Fatalf("CreateChannel failed: %v", err)
	}

	if channel.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, channel.Name)
	}
	if channel.MaxUsers != MaxUsersPerChannel {
		t.Errorf("Expected max users %d, got %d", MaxUsersPerChannel, channel.MaxUsers)
	}
}

func TestChannelService_JoinChannel(t *testing.T) {
	service := NewChannelService(NewMemoryChannelStore())

	channel, _ := service.CreateChannel("user1", &CreateChannelRequest{
		Name:      "Test Channel",
		Latitude:  -23.5505,
		Longitude: -46.6333,
	})

	err := service.JoinChannel("user2", channel.ID)
	if err != nil {
		t.Fatalf("JoinChannel failed: %v", err)
	}

	updated, _ := service.GetChannel(channel.ID)
	if updated.UserCount != 2 {
		t.Errorf("Expected 2 users (creator + user2), got %d", updated.UserCount)
	}
}

func TestChannelService_JoinChannel_AlreadyInChannel(t *testing.T) {
	service := NewChannelService(NewMemoryChannelStore())

	channel, _ := service.CreateChannel("user1", &CreateChannelRequest{
		Name:      "Test Channel",
		Latitude:  -23.5505,
		Longitude: -46.6333,
	})

	service.JoinChannel("user2", channel.ID)

	err := service.JoinChannel("user2", channel.ID)
	if err != ErrUserAlreadyInChannel {
		t.Errorf("Expected ErrUserAlreadyInChannel, got %v", err)
	}
}

func TestChannelService_LeaveChannel(t *testing.T) {
	service := NewChannelService(NewMemoryChannelStore())

	channel, _ := service.CreateChannel("user1", &CreateChannelRequest{
		Name:      "Test Channel",
		Latitude:  -23.5505,
		Longitude: -46.6333,
	})

	service.JoinChannel("user2", channel.ID)

	err := service.LeaveChannel("user2", channel.ID)
	if err != nil {
		t.Fatalf("LeaveChannel failed: %v", err)
	}

	updated, _ := service.GetChannel(channel.ID)
	if updated.UserCount != 1 {
		t.Errorf("Expected 1 user (only the creator), got %d", updated.UserCount)
	}
}

func TestChannelService_ChannelFull(t *testing.T) {
	service := NewChannelService(NewMemoryChannelStore())

	channel, _ := service.CreateChannel("user1", &CreateChannelRequest{
		Name:      "Test Channel",
		Latitude:  -23.5505,
		Longitude: -46.6333,
	})

	// The creator is already a member, so only MaxUsersPerChannel-1 others fit.
	for i := 2; i <= MaxUsersPerChannel; i++ {
		service.JoinChannel(fmt.Sprintf("user%d", i), channel.ID)
	}

	// Verify channel is full
	ch, _ := service.GetChannel(channel.ID)
	if ch.UserCount != MaxUsersPerChannel {
		t.Fatalf("Expected channel to be full with %d users, got %d", MaxUsersPerChannel, ch.UserCount)
	}

	// Try to join a full channel
	err := service.JoinChannel("userExtra", channel.ID)
	if err != ErrChannelFull {
		t.Errorf("Expected ErrChannelFull, got %v", err)
	}
}

func TestChannelService_GetNearbyChannels(t *testing.T) {
	service := NewChannelService(NewMemoryChannelStore())

	service.CreateChannel("user1", &CreateChannelRequest{
		Name:      "Channel 1",
		Latitude:  -23.5505,
		Longitude: -46.6333,
	})

	service.CreateChannel("user2", &CreateChannelRequest{
		Name:      "Channel 2",
		Latitude:  -22.9068,
		Longitude: -43.1729,
	})

	// Use a larger radius to capture both channels
	channels, err := service.GetNearbyChannels(-23.5505, -46.6333, 500.0)
	if err != nil {
		t.Fatalf("GetNearbyChannels failed: %v", err)
	}

	if len(channels) < 1 {
		t.Errorf("Expected at least 1 nearby channel, got %d", len(channels))
	}
}

func TestChannelService_MaxVisibleChannels(t *testing.T) {
	service := NewChannelService(NewMemoryChannelStore())

	for i := 0; i < 20; i++ {
		service.CreateChannel("user"+string(rune('0'+i)), &CreateChannelRequest{
			Name:      "Channel " + string(rune('A'+i)),
			Latitude:  -23.5505,
			Longitude: -46.6333,
		})
	}

	channels, _ := service.GetNearbyChannels(-23.5505, -46.6333, 100.0)

	if len(channels) > MaxVisibleChannels {
		t.Errorf("Expected max %d channels, got %d", MaxVisibleChannels, len(channels))
	}
}

// A4 / RF05: quando todos os canais proximos estao cheios, o usuario excedente
// deve ser redirecionado para um canal novo (JoinChannelOrCreate) em vez de
// falhar, e o canal cheio permanece com exatamente MaxUsersPerChannel membros.
func TestChannelService_JoinChannelOrCreate_FullRedirectsToNewChannel(t *testing.T) {
	service := NewChannelService(NewMemoryChannelStore())
	req := &CreateChannelRequest{
		Name:      "RF05 Channel",
		Latitude:  -23.5505,
		Longitude: -46.6333,
	}

	first, err := service.CreateChannel("user1", req)
	if err != nil {
		t.Fatalf("CreateChannel: %v", err)
	}
	for i := 2; i <= MaxUsersPerChannel; i++ {
		if err := service.JoinChannel(fmt.Sprintf("user%d", i), first.ID); err != nil {
			t.Fatalf("fill join %d: %v", i, err)
		}
	}
	full, err := service.GetChannel(first.ID)
	if err != nil || full.UserCount != MaxUsersPerChannel {
		t.Fatalf("setup: channel should be full, got user_count=%d err=%v", full.UserCount, err)
	}

	joined, err := service.JoinChannelOrCreate("user11", req)
	if err != nil {
		t.Fatalf("JoinChannelOrCreate: %v", err)
	}
	if joined.ID == first.ID {
		t.Error("expected redirect to a new channel, got the full one")
	}

	after, err := service.GetChannel(first.ID)
	if err != nil {
		t.Fatalf("GetChannel(full): %v", err)
	}
	if after.UserCount != MaxUsersPerChannel {
		t.Errorf("full channel must stay at %d members, got %d", MaxUsersPerChannel, after.UserCount)
	}

	members, err := service.GetChannelUsers(joined.ID)
	if err != nil {
		t.Fatalf("GetChannelUsers: %v", err)
	}
	found := false
	for _, m := range members {
		if m == "user11" {
			found = true
		}
	}
	if !found {
		t.Errorf("user11 should be a member of the new channel, members=%v", members)
	}
}

// B2 / RF05: joins concorrentes nunca estouram o teto de 10 por canal - ou o
// candidato aceita, ou ErrChannelFull leva a cria/candidato seguinte.
func TestChannelService_JoinChannelOrCreate_ConcurrentNeverExceedsMax(t *testing.T) {
	service := NewChannelService(NewMemoryChannelStore())
	req := &CreateChannelRequest{
		Name:      "RF05 Concurrent",
		Latitude:  -23.5505,
		Longitude: -46.6333,
	}

	const users = 25
	var wg sync.WaitGroup
	joinedID := make([]string, users)
	errs := make([]error, users)
	for i := 0; i < users; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ch, err := service.JoinChannelOrCreate(fmt.Sprintf("conc-user-%d", i), req)
			if err == nil && ch != nil {
				joinedID[i] = ch.ID
			}
			errs[i] = err
		}(i)
	}
	wg.Wait()

	seen := make(map[string]bool)
	for i := 0; i < users; i++ {
		if errs[i] != nil {
			t.Errorf("join %d: %v", i, errs[i])
			continue
		}
		if joinedID[i] == "" {
			t.Errorf("join %d: no channel returned", i)
			continue
		}
		seen[joinedID[i]] = true
	}

	for id := range seen {
		ch, err := service.GetChannel(id)
		if err != nil {
			t.Errorf("GetChannel(%s): %v", id, err)
			continue
		}
		if ch.UserCount > MaxUsersPerChannel {
			t.Errorf("channel %s exceeded cap: %d > %d", id, ch.UserCount, MaxUsersPerChannel)
		}
	}
	if len(seen) < 2 {
		t.Errorf("expected at least 2 channels for %d users (cap %d each), got %d", users, MaxUsersPerChannel, len(seen))
	}
}
