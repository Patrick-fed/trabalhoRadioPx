package location

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

var ErrUserLocationNotFound = errors.New("user location not found")

const (
	EarthRadiusKm = 6371.0
	DefaultRadiusKm = 10.0
	MaxVisibleChannels = 15
	MaxUsersPerChannel = 10
)

type LocationService struct {
	mu              sync.RWMutex
	userLocations   map[string]*Location
	channelLocations map[string]*Location
}

func NewLocationService() *LocationService {
	return &LocationService{
		userLocations:   make(map[string]*Location),
		channelLocations: make(map[string]*Location),
	}
}

func (s *LocationService) UpdateUserLocation(userID string, update *LocationUpdate) (*Location, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	location, exists := s.userLocations[userID]
	if exists {
		location.Latitude = update.Latitude
		location.Longitude = update.Longitude
		location.Accuracy = update.Accuracy
		location.Timestamp = now
		location.UpdatedAt = now
	} else {
		location = &Location{
			ID:        generateID(),
			UserID:    userID,
			Latitude:  update.Latitude,
			Longitude: update.Longitude,
			Accuracy:  update.Accuracy,
			Timestamp: now,
			UpdatedAt: now,
		}
		s.userLocations[userID] = location
	}

	return location, nil
}

func (s *LocationService) GetUserLocation(userID string) (*Location, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	location, exists := s.userLocations[userID]
	return location, exists
}

func (s *LocationService) UpdateChannelLocation(channelID string, location *Location) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.channelLocations[channelID] = location
}

func (s *LocationService) FindNearbyChannels(userID string, filter *ProximityFilter) ([]*NearbyChannel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userLocation, exists := s.userLocations[userID]
	if !exists {
		return nil, ErrUserLocationNotFound
	}

	radiusKm := filter.RadiusKm
	if radiusKm == 0 {
		radiusKm = DefaultRadiusKm
	}

	maxResults := filter.MaxResults
	if maxResults == 0 {
		maxResults = MaxVisibleChannels
	}

	var nearbyChannels []*NearbyChannel

	for channelID, channelLocation := range s.channelLocations {
		distance := CalculateDistance(
			userLocation.Latitude, userLocation.Longitude,
			channelLocation.Latitude, channelLocation.Longitude,
		)

		if distance <= radiusKm {
			nearbyChannel := &NearbyChannel{
				ChannelID:  channelID,
				Name:       getChannelName(channelID),
				Latitude:   channelLocation.Latitude,
				Longitude:  channelLocation.Longitude,
				Distance:   distance,
				UserCount:  getChannelUserCount(channelID),
				MaxUsers:   MaxUsersPerChannel,
				IsAvailable: getChannelUserCount(channelID) < MaxUsersPerChannel,
			}
			nearbyChannels = append(nearbyChannels, nearbyChannel)
		}
	}

	sortNearbyChannels(nearbyChannels)

	if len(nearbyChannels) > maxResults {
		nearbyChannels = nearbyChannels[:maxResults]
	}

	return nearbyChannels, nil
}

func (s *LocationService) GetUsersInRadius(latitude, longitude, radiusKm float64) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var users []string

	for userID, location := range s.userLocations {
		distance := CalculateDistance(
			latitude, longitude,
			location.Latitude, location.Longitude,
		)

		if distance <= radiusKm {
			users = append(users, userID)
		}
	}

	return users
}

func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadiusKm * c
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func getChannelName(channelID string) string {
	return "Channel " + channelID[:8]
}

func getChannelUserCount(channelID string) int {
	return 0
}

func sortNearbyChannels(channels []*NearbyChannel) {
	for i := 0; i < len(channels)-1; i++ {
		for j := i + 1; j < len(channels); j++ {
			if channels[j].Distance < channels[i].Distance {
				channels[i], channels[j] = channels[j], channels[i]
			}
		}
	}
}
