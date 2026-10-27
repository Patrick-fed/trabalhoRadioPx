package channel

import (
	"sync"
	"time"
)

func now() time.Time {
	return time.Now()
}

// MemoryChannelStore keeps channels and memberships in RAM. It backs the unit
// tests and keeps the service testable without a PostgreSQL instance.
type MemoryChannelStore struct {
	mu           sync.RWMutex
	channels     map[string]*Channel
	channelUsers map[string][]string
}

func NewMemoryChannelStore() *MemoryChannelStore {
	return &MemoryChannelStore{
		channels:     make(map[string]*Channel),
		channelUsers: make(map[string][]string),
	}
}

func (s *MemoryChannelStore) Create(ch *Channel) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.channels[ch.ID] = ch
	s.channelUsers[ch.ID] = []string{ch.CreatedBy}
	return nil
}

func (s *MemoryChannelStore) GetByID(id string) (*Channel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ch, exists := s.channels[id]
	if !exists {
		return nil, ErrChannelNotFound
	}

	chCopy := *ch
	chCopy.UserCount = len(s.channelUsers[id])
	chCopy.IsAvailable = chCopy.UserCount < chCopy.MaxUsers
	return &chCopy, nil
}

func (s *MemoryChannelStore) AddMember(channelID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch, exists := s.channels[channelID]
	if !exists {
		return ErrChannelNotFound
	}

	for _, u := range s.channelUsers[channelID] {
		if u == userID {
			return ErrUserAlreadyInChannel
		}
	}

	if len(s.channelUsers[channelID]) >= ch.MaxUsers {
		return ErrChannelFull
	}

	s.channelUsers[channelID] = append(s.channelUsers[channelID], userID)
	ch.UserCount = len(s.channelUsers[channelID])
	ch.IsAvailable = ch.UserCount < ch.MaxUsers
	ch.UpdatedAt = now()
	return nil
}

func (s *MemoryChannelStore) RemoveMember(channelID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch, exists := s.channels[channelID]
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

	ch.UserCount = len(s.channelUsers[channelID])
	ch.IsAvailable = ch.UserCount < ch.MaxUsers
	ch.UpdatedAt = now()
	return nil
}

func (s *MemoryChannelStore) IsMember(channelID, userID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.channels[channelID]; !exists {
		return false, ErrChannelNotFound
	}

	for _, u := range s.channelUsers[channelID] {
		if u == userID {
			return true, nil
		}
	}
	return false, nil
}

func (s *MemoryChannelStore) ListMembers(channelID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users, exists := s.channelUsers[channelID]
	if !exists {
		return nil, ErrChannelNotFound
	}

	members := make([]string, len(users))
	copy(members, users)
	return members, nil
}

func (s *MemoryChannelStore) ListUserChannels(userID string) ([]*Channel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Channel, 0)
	for channelID, users := range s.channelUsers {
		for _, u := range users {
			if u == userID {
				if ch, exists := s.channels[channelID]; exists {
					chCopy := *ch
					chCopy.UserCount = len(users)
					chCopy.IsAvailable = chCopy.UserCount < chCopy.MaxUsers
					result = append(result, &chCopy)
				}
				break
			}
		}
	}
	return result, nil
}

// ListNearby returns a copy of every channel (the bounding-box pre-filter is a
// PostgreSQL optimization; in memory the full scan is cheap). Distance
// filtering and sorting are the service's job.
func (s *MemoryChannelStore) ListNearby(latitude, longitude, radiusKm float64) ([]*Channel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Channel, 0, len(s.channels))
	for _, ch := range s.channels {
		chCopy := *ch
		chCopy.UserCount = len(s.channelUsers[ch.ID])
		chCopy.IsAvailable = chCopy.UserCount < chCopy.MaxUsers
		result = append(result, &chCopy)
	}
	return result, nil
}

func (s *MemoryChannelStore) Update(ch *Channel) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.channels[ch.ID]; !exists {
		return ErrChannelNotFound
	}
	s.channels[ch.ID] = ch
	return nil
}

func (s *MemoryChannelStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.channels[id]; !exists {
		return ErrChannelNotFound
	}
	delete(s.channels, id)
	delete(s.channelUsers, id)
	return nil
}
