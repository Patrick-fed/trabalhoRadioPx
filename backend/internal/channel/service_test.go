package channel

import (
	"fmt"
	"testing"

	"github.com/radiopx/backend/internal/location"
)

func TestChannelService_CreateChannel(t *testing.T) {
	locationService := location.NewLocationService()
	service := NewChannelService(locationService)

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
	locationService := location.NewLocationService()
	service := NewChannelService(locationService)

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
	if updated.UserCount != 1 {
		t.Errorf("Expected 1 user, got %d", updated.UserCount)
	}
}

func TestChannelService_JoinChannel_AlreadyInChannel(t *testing.T) {
	locationService := location.NewLocationService()
	service := NewChannelService(locationService)

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
	locationService := location.NewLocationService()
	service := NewChannelService(locationService)

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
	if updated.UserCount != 0 {
		t.Errorf("Expected 0 users, got %d", updated.UserCount)
	}
}

func TestChannelService_ChannelFull(t *testing.T) {
	locationService := location.NewLocationService()
	service := NewChannelService(locationService)

	channel, _ := service.CreateChannel("user1", &CreateChannelRequest{
		Name:      "Test Channel",
		Latitude:  -23.5505,
		Longitude: -46.6333,
	})

	// Fill the channel to capacity
	for i := 1; i <= MaxUsersPerChannel; i++ {
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
	locationService := location.NewLocationService()
	service := NewChannelService(locationService)

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
	locationService := location.NewLocationService()
	service := NewChannelService(locationService)

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
