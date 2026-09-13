package channel

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/radiopx/backend/internal/location"
)

const (
	MaxUsersPerChannel = 10
	MaxVisibleChannels = 15
)

var (
	ErrChannelNotFound      = errors.New("channel not found")
	ErrChannelFull          = errors.New("channel is full")
	ErrChannelAlreadyExists = errors.New("channel already exists")
	ErrUserAlreadyInChannel = errors.New("user already in channel")
	ErrUserNotInChannel     = errors.New("user not in channel")
	ErrMaxChannelsReached   = errors.New("maximum visible channels reached")
)

type ChannelService struct {
	mu           sync.RWMutex
	channels     map[string]*Channel
	channelUsers map[string][]string
	locationService *location.LocationService
}

func NewChannelService(locationService *location.LocationService) *ChannelService {
	return &ChannelService{
		channels:        make(map[string]*Channel),
		channelUsers:    make(map[string][]string),
		locationService: locationService,
	}
}

func (s *ChannelService) CreateChannel(userID string, req *CreateChannelRequest) (*Channel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	channel := &Channel{
		ID:          generateChannelID(),
		Name:        req.Name,
		Description: req.Description,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		MaxUsers:    MaxUsersPerChannel,
		UserCount:   0,
		IsAvailable: true,
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.channels[channel.ID] = channel

	s.locationService.UpdateChannelLocation(channel.ID, &location.Location{
		ID:        channel.ID,
		Latitude:  channel.Latitude,
		Longitude: channel.Longitude,
		Timestamp: now,
		UpdatedAt: now,
	})

	return channel, nil
}

func (s *ChannelService) GetChannel(channelID string) (*Channel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	channel, exists := s.channels[channelID]
	if !exists {
		return nil, ErrChannelNotFound
	}

	return channel, nil
}

func (s *ChannelService) JoinChannel(userID, channelID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	channel, exists := s.channels[channelID]
	if !exists {
		return ErrChannelNotFound
	}

	if channel.UserCount >= channel.MaxUsers {
		return ErrChannelFull
	}

	users := s.channelUsers[channelID]
	for _, u := range users {
		if u == userID {
			return ErrUserAlreadyInChannel
		}
	}

	s.channelUsers[channelID] = append(users, userID)
	channel.UserCount++
	channel.IsAvailable = channel.UserCount < channel.MaxUsers
	channel.UpdatedAt = time.Now()

	return nil
}

func (s *ChannelService) JoinChannelOrCreate(userID string, req *CreateChannelRequest) (*Channel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nearbyChannels := make([]*Channel, 0)
	for _, channel := range s.channels {
		distance := location.CalculateDistance(
			req.Latitude, req.Longitude,
			channel.Latitude, channel.Longitude,
		)
		if distance <= 10.0 && channel.IsAvailable {
			nearbyChannels = append(nearbyChannels, channel)
		}
	}

	if len(nearbyChannels) > 0 {
		for _, channel := range nearbyChannels {
			users := s.channelUsers[channel.ID]
			alreadyInChannel := false
			for _, u := range users {
				if u == userID {
					alreadyInChannel = true
					break
				}
			}

			if !alreadyInChannel && channel.UserCount < channel.MaxUsers {
				s.channelUsers[channel.ID] = append(users, userID)
				channel.UserCount++
				channel.IsAvailable = channel.UserCount < channel.MaxUsers
				channel.UpdatedAt = time.Now()
				return channel, nil
			}
		}
	}

	now := time.Now()
	channel := &Channel{
		ID:          generateChannelID(),
		Name:        req.Name,
		Description: req.Description,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		MaxUsers:    MaxUsersPerChannel,
		UserCount:   1,
		IsAvailable: true,
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.channels[channel.ID] = channel
	s.channelUsers[channel.ID] = []string{userID}

	s.locationService.UpdateChannelLocation(channel.ID, &location.Location{
		ID:        channel.ID,
		Latitude:  channel.Latitude,
		Longitude: channel.Longitude,
		Timestamp: now,
		UpdatedAt: now,
	})

	return channel, nil
}

func (s *ChannelService) LeaveChannel(userID, channelID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	channel, exists := s.channels[channelID]
	if !exists {
		return ErrChannelNotFound
	}

	users := s.channelUsers[channelID]
	found := false
	for i, u := range users {
		if u == userID {
			s.channelUsers[channelID] = append(users[:i], users[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return ErrUserNotInChannel
	}

	channel.UserCount = len(s.channelUsers[channelID])
	channel.IsAvailable = channel.UserCount < channel.MaxUsers
	channel.UpdatedAt = time.Now()

	return nil
}

func (s *ChannelService) GetNearbyChannels(latitude, longitude, radiusKm float64) ([]*Channel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nearbyChannels := make([]*Channel, 0)

	for _, channel := range s.channels {
		distance := location.CalculateDistance(
			latitude, longitude,
			channel.Latitude, channel.Longitude,
		)

		if distance <= radiusKm {
			channelCopy := *channel
			nearbyChannels = append(nearbyChannels, &channelCopy)
		}
	}

	sortChannelsByDistance(nearbyChannels, latitude, longitude)

	if len(nearbyChannels) > MaxVisibleChannels {
		nearbyChannels = nearbyChannels[:MaxVisibleChannels]
	}

	return nearbyChannels, nil
}

func (s *ChannelService) GetUserChannels(userID string) []*Channel {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var userChannels []*Channel

	for channelID, users := range s.channelUsers {
		for _, u := range users {
			if u == userID {
				if channel, exists := s.channels[channelID]; exists {
					channelCopy := *channel
					userChannels = append(userChannels, &channelCopy)
				}
				break
			}
		}
	}

	return userChannels
}

func (s *ChannelService) GetChannelUsers(channelID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users, exists := s.channelUsers[channelID]
	if !exists {
		return nil, ErrChannelNotFound
	}

	usersCopy := make([]string, len(users))
	copy(usersCopy, users)

	return usersCopy, nil
}

func (s *ChannelService) UpdateChannel(channelID, userID string, req *UpdateChannelRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	channel, exists := s.channels[channelID]
	if !exists {
		return ErrChannelNotFound
	}

	if channel.CreatedBy != userID {
		return errors.New("only channel creator can update")
	}

	if req.Name != "" {
		channel.Name = req.Name
	}
	if req.Description != "" {
		channel.Description = req.Description
	}
	channel.UpdatedAt = time.Now()

	return nil
}

func (s *ChannelService) DeleteChannel(channelID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	channel, exists := s.channels[channelID]
	if !exists {
		return ErrChannelNotFound
	}

	if channel.CreatedBy != userID {
		return errors.New("only channel creator can delete")
	}

	delete(s.channels, channelID)
	delete(s.channelUsers, channelID)

	return nil
}

func generateChannelID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func sortChannelsByDistance(channels []*Channel, userLat, userLng float64) {
	for i := 0; i < len(channels)-1; i++ {
		for j := i + 1; j < len(channels); j++ {
			distI := location.CalculateDistance(userLat, userLng, channels[i].Latitude, channels[i].Longitude)
			distJ := location.CalculateDistance(userLat, userLng, channels[j].Latitude, channels[j].Longitude)
			if distJ < distI {
				channels[i], channels[j] = channels[j], channels[i]
			}
		}
	}
}
