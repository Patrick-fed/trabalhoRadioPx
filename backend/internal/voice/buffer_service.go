package voice

import (
	"fmt"
	"sync"
	"time"
)

type AudioBufferService struct {
	mu          sync.RWMutex
	buffers     map[string][]*MessageBuffer
	config      *BufferConfig
	sequenceMap map[string]int64
}

func NewAudioBufferService(config *BufferConfig) *AudioBufferService {
	if config == nil {
		config = DefaultBufferConfig()
	}

	return &AudioBufferService{
		buffers:     make(map[string][]*MessageBuffer),
		config:      config,
		sequenceMap: make(map[string]int64),
	}
}

func (s *AudioBufferService) AddMessage(channelID, userID string, audioData []byte) (*MessageBuffer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sequence := s.sequenceMap[channelID] + 1
	s.sequenceMap[channelID] = sequence

	now := time.Now()
	message := &MessageBuffer{
		ID:        generateBufferID(),
		ChannelID: channelID,
		UserID:    userID,
		AudioData: audioData,
		Sequence:  sequence,
		Timestamp: now,
		CreatedAt: now,
	}

	s.buffers[channelID] = append(s.buffers[channelID], message)

	s.cleanupOldMessages(channelID)

	return message, nil
}

func (s *AudioBufferService) GetMessages(channelID string, sinceSequence int64) []*MessageBuffer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messages := s.buffers[channelID]
	var result []*MessageBuffer

	for _, msg := range messages {
		if msg.Sequence > sinceSequence && !msg.IsReplayed {
			result = append(result, msg)
		}
	}

	return result
}

func (s *AudioBufferService) GetUnreplayedMessages(channelID string) []*MessageBuffer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messages := s.buffers[channelID]
	var result []*MessageBuffer

	for _, msg := range messages {
		if !msg.IsReplayed {
			result = append(result, msg)
		}
	}

	return result
}

func (s *AudioBufferService) MarkAsReplayed(channelID string, messageIDs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	messages := s.buffers[channelID]
	idSet := make(map[string]bool)
	for _, id := range messageIDs {
		idSet[id] = true
	}

	for _, msg := range messages {
		if idSet[msg.ID] {
			msg.IsReplayed = true
		}
	}
}

func (s *AudioBufferService) MarkAllAsReplayed(channelID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	messages := s.buffers[channelID]
	for _, msg := range messages {
		msg.IsReplayed = true
	}
}

func (s *AudioBufferService) GetStatus() *BufferStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalMessages := 0
	var totalSize int64
	var oldest, newest *time.Time
	channelCount := len(s.buffers)

	for _, messages := range s.buffers {
		for _, msg := range messages {
			totalMessages++
			totalSize += int64(len(msg.AudioData))

			if oldest == nil || msg.Timestamp.Before(*oldest) {
				oldest = &msg.Timestamp
			}
			if newest == nil || msg.Timestamp.After(*newest) {
				newest = &msg.Timestamp
			}
		}
	}

	return &BufferStatus{
		TotalMessages:  totalMessages,
		TotalSizeBytes: totalSize,
		OldestMessage:  oldest,
		NewestMessage:  newest,
		ChannelCount:   channelCount,
	}
}

func (s *AudioBufferService) cleanupOldMessages(channelID string) {
	messages := s.buffers[channelID]
	now := time.Now()
	cutoff := now.Add(-s.config.MaxDuration)

	var cleaned []*MessageBuffer
	for _, msg := range messages {
		if msg.Timestamp.After(cutoff) {
			cleaned = append(cleaned, msg)
		}
	}

	s.buffers[channelID] = cleaned

	var totalSize int64
	for _, msg := range cleaned {
		totalSize += int64(len(msg.AudioData))
	}

	if totalSize > s.config.MaxSizeBytes {
		s.cleanupBySize(channelID)
	}
}

func (s *AudioBufferService) cleanupBySize(channelID string) {
	messages := s.buffers[channelID]
	if len(messages) == 0 {
		return
	}

	var totalSize int64
	for _, msg := range messages {
		totalSize += int64(len(msg.AudioData))
	}

	for totalSize > s.config.MaxSizeBytes && len(messages) > 1 {
		totalSize -= int64(len(messages[0].AudioData))
		messages = messages[1:]
	}

	s.buffers[channelID] = messages
}

func (s *AudioBufferService) CleanupAllChannels() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for channelID := range s.buffers {
		s.cleanupOldMessages(channelID)
	}
}

func (s *AudioBufferService) ClearChannel(channelID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.buffers, channelID)
	delete(s.sequenceMap, channelID)
}

func (s *AudioBufferService) StartCleanupRoutine() {
	ticker := time.NewTicker(s.config.CleanupInterval)
	go func() {
		for range ticker.C {
			s.CleanupAllChannels()
		}
	}()
}

func generateBufferID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
