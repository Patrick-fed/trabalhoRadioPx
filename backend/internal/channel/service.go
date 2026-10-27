package channel

import (
	"errors"
	"time"

	"github.com/google/uuid"
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

// ChannelService holds channel business rules. Persistence goes through
// ChannelStore (PostgreSQL in production, in-memory for the unit tests);
// UserCount is always computed by the store, never incremented here.
type ChannelService struct {
	store ChannelStore
}

func NewChannelService(store ChannelStore) *ChannelService {
	return &ChannelService{store: store}
}

func (s *ChannelService) CreateChannel(userID string, req *CreateChannelRequest) (*Channel, error) {
	now := time.Now()
	ch := &Channel{
		ID:          generateChannelID(),
		Name:        req.Name,
		Description: req.Description,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		MaxUsers:    MaxUsersPerChannel,
		UserCount:   1,
		IsAvailable: 1 < MaxUsersPerChannel,
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.store.Create(ch); err != nil {
		return nil, err
	}

	return ch, nil
}

func (s *ChannelService) GetChannel(channelID string) (*Channel, error) {
	return s.store.GetByID(channelID)
}

func (s *ChannelService) JoinChannel(userID, channelID string) error {
	if _, err := s.store.GetByID(channelID); err != nil {
		return err
	}
	return s.store.AddMember(channelID, userID)
}

func (s *ChannelService) JoinChannelOrCreate(userID string, req *CreateChannelRequest) (*Channel, error) {
	const joinRadiusKm = 10.0

	candidates, err := s.store.ListNearby(req.Latitude, req.Longitude, joinRadiusKm)
	if err != nil {
		return nil, err
	}

	for _, ch := range candidates {
		distance := location.CalculateDistance(
			req.Latitude, req.Longitude,
			ch.Latitude, ch.Longitude,
		)
		if distance > joinRadiusKm || !ch.IsAvailable {
			continue
		}

		isMember, err := s.store.IsMember(ch.ID, userID)
		if err != nil || isMember {
			continue
		}

		if err := s.store.AddMember(ch.ID, userID); err != nil {
			continue
		}
		return s.store.GetByID(ch.ID)
	}

	return s.CreateChannel(userID, req)
}

func (s *ChannelService) LeaveChannel(userID, channelID string) error {
	return s.store.RemoveMember(channelID, userID)
}

func (s *ChannelService) GetNearbyChannels(latitude, longitude, radiusKm float64) ([]*Channel, error) {
	candidates, err := s.store.ListNearby(latitude, longitude, radiusKm)
	if err != nil {
		return nil, err
	}

	nearbyChannels := make([]*Channel, 0, len(candidates))
	for _, ch := range candidates {
		distance := location.CalculateDistance(
			latitude, longitude,
			ch.Latitude, ch.Longitude,
		)
		if distance <= radiusKm {
			chCopy := *ch
			chCopy.Distance = distance
			nearbyChannels = append(nearbyChannels, &chCopy)
		}
	}

	sortChannelsByDistance(nearbyChannels, latitude, longitude)

	if len(nearbyChannels) > MaxVisibleChannels {
		nearbyChannels = nearbyChannels[:MaxVisibleChannels]
	}

	return nearbyChannels, nil
}

func (s *ChannelService) GetUserChannels(userID string) ([]*Channel, error) {
	return s.store.ListUserChannels(userID)
}

func (s *ChannelService) GetChannelUsers(channelID string) ([]string, error) {
	return s.store.ListMembers(channelID)
}

func (s *ChannelService) UpdateChannel(channelID, userID string, req *UpdateChannelRequest) error {
	ch, err := s.store.GetByID(channelID)
	if err != nil {
		return err
	}

	if ch.CreatedBy != userID {
		return errors.New("only channel creator can update")
	}

	if req.Name != "" {
		ch.Name = req.Name
	}
	if req.Description != "" {
		ch.Description = req.Description
	}
	ch.UpdatedAt = time.Now()

	return s.store.Update(ch)
}

func (s *ChannelService) DeleteChannel(channelID, userID string) error {
	ch, err := s.store.GetByID(channelID)
	if err != nil {
		return err
	}

	if ch.CreatedBy != userID {
		return errors.New("only channel creator can delete")
	}

	return s.store.Delete(channelID)
}

func generateChannelID() string {
	return uuid.NewString()
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
