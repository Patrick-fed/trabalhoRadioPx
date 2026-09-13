package voice

import (
	"log"
	"sync"
	"time"
)

type VoiceService struct {
	hub            *VoiceHub
	mu             sync.RWMutex
}

type VoiceHub struct {
	channels map[string]*ChannelState
	mu       sync.RWMutex
}

type ChannelState struct {
	ChannelID      string
	CurrentSpeaker string
	IsTransmitting bool
	LastActivity   time.Time
}

func NewVoiceService() *VoiceService {
	return &VoiceService{
		hub: &VoiceHub{
			channels: make(map[string]*ChannelState),
		},
	}
}

func (vs *VoiceService) StartTransmission(userID, channelID string) error {
	vs.hub.mu.Lock()
	defer vs.hub.mu.Unlock()

	channel, exists := vs.hub.channels[channelID]
	if !exists {
		channel = &ChannelState{
			ChannelID: channelID,
		}
		vs.hub.channels[channelID] = channel
	}

	if channel.IsTransmitting && channel.CurrentSpeaker != userID {
		log.Printf("Channel %s is busy, speaker: %s", channelID, channel.CurrentSpeaker)
		return ErrChannelBusy
	}

	channel.CurrentSpeaker = userID
	channel.IsTransmitting = true
	channel.LastActivity = time.Now()

	log.Printf("User %s started transmission in channel %s", userID, channelID)
	return nil
}

func (vs *VoiceService) StopTransmission(userID, channelID string) error {
	vs.hub.mu.Lock()
	defer vs.hub.mu.Unlock()

	channel, exists := vs.hub.channels[channelID]
	if !exists {
		return ErrChannelNotFound
	}

	if channel.CurrentSpeaker != userID {
		return ErrNotSpeaker
	}

	channel.CurrentSpeaker = ""
	channel.IsTransmitting = false
	channel.LastActivity = time.Now()

	log.Printf("User %s stopped transmission in channel %s", userID, channelID)
	return nil
}

func (vs *VoiceService) IsChannelBusy(channelID string) bool {
	vs.hub.mu.RLock()
	defer vs.hub.mu.RUnlock()

	channel, exists := vs.hub.channels[channelID]
	if !exists {
		return false
	}

	return channel.IsTransmitting
}

func (vs *VoiceService) GetCurrentSpeaker(channelID string) string {
	vs.hub.mu.RLock()
	defer vs.hub.mu.RUnlock()

	channel, exists := vs.hub.channels[channelID]
	if !exists {
		return ""
	}

	return channel.CurrentSpeaker
}

func (vs *VoiceService) GetChannelStatus(channelID string) *VoiceStatus {
	vs.hub.mu.RLock()
	defer vs.hub.mu.RUnlock()

	channel, exists := vs.hub.channels[channelID]
	if !exists {
		return &VoiceStatus{
			IsTransmitting: false,
			ChannelID:      channelID,
		}
	}

	return &VoiceStatus{
		IsTransmitting: channel.IsTransmitting,
		CurrentSpeaker: channel.CurrentSpeaker,
		ChannelID:      channelID,
	}
}

func (vs *VoiceService) CleanupInactiveChannels(timeout time.Duration) {
	vs.hub.mu.Lock()
	defer vs.hub.mu.Unlock()

	now := time.Now()
	for channelID, channel := range vs.hub.channels {
		if now.Sub(channel.LastActivity) > timeout {
			delete(vs.hub.channels, channelID)
			log.Printf("Cleaned up inactive channel: %s", channelID)
		}
	}
}

var (
	ErrChannelBusy     = &VoiceError{Code: "CHANNEL_BUSY", Message: "Channel is currently in use"}
	ErrChannelNotFound = &VoiceError{Code: "CHANNEL_NOT_FOUND", Message: "Channel not found"}
	ErrNotSpeaker      = &VoiceError{Code: "NOT_SPEAKER", Message: "User is not the current speaker"}
)

type VoiceError struct {
	Code    string
	Message string
}

func (e *VoiceError) Error() string {
	return e.Message
}
