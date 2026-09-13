package voice

import (
	"sync"
	"time"
)

type BufferManager struct {
	mu              sync.RWMutex
	bufferService   *AudioBufferService
	channelConfigs  map[string]*BufferConfig
	reconnectStates map[string]*ReconnectState
}

type ReconnectState struct {
	LastSequence    int64
	LastReconnect   time.Time
	PendingMessages []*MessageBuffer
	IsReplaying     bool
}

func NewBufferManager(bufferService *AudioBufferService) *BufferManager {
	return &BufferManager{
		bufferService:   bufferService,
		channelConfigs:  make(map[string]*BufferConfig),
		reconnectStates: make(map[string]*ReconnectState),
	}
}

func (m *BufferManager) ConfigureChannel(channelID string, config *BufferConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if config == nil {
		config = DefaultBufferConfig()
	}
	m.channelConfigs[channelID] = config
}

func (m *BufferManager) GetChannelConfig(channelID string) *BufferConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config, exists := m.channelConfigs[channelID]
	if !exists {
		return DefaultBufferConfig()
	}
	return config
}

func (m *BufferManager) HandleDisconnection(channelID, userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.reconnectStates[channelID]
	if !exists {
		state = &ReconnectState{}
		m.reconnectStates[channelID] = state
	}

	state.LastReconnect = time.Now()
}

func (m *BufferManager) HandleReconnection(channelID, userID string) []*MessageBuffer {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.reconnectStates[channelID]
	if !exists {
		return nil
	}

	if state.IsReplaying {
		return nil
	}

	state.IsReplaying = true

	messages := m.bufferService.GetUnreplayedMessages(channelID)
	state.PendingMessages = messages

	return messages
}

func (m *BufferManager) ConfirmReplay(channelID string, messageIDs []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.bufferService.MarkAsReplayed(channelID, messageIDs)

	state, exists := m.reconnectStates[channelID]
	if exists {
		state.IsReplaying = false
		state.PendingMessages = nil
		state.LastSequence = m.getLastSequence(channelID)
	}
}

func (m *BufferManager) CompleteReplay(channelID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.bufferService.MarkAllAsReplayed(channelID)

	state, exists := m.reconnectStates[channelID]
	if exists {
		state.IsReplaying = false
		state.PendingMessages = nil
		state.LastSequence = m.getLastSequence(channelID)
	}
}

func (m *BufferManager) GetReconnectState(channelID string) *ReconnectState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.reconnectStates[channelID]
	if !exists {
		return &ReconnectState{}
	}
	return state
}

func (m *BufferManager) getLastSequence(channelID string) int64 {
	messages := m.bufferService.GetMessages(channelID, 0)
	if len(messages) == 0 {
		return 0
	}
	return messages[len(messages)-1].Sequence
}

func (m *BufferManager) GetBufferStatus(channelID string) map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := m.bufferService.GetStatus()
	config := m.GetChannelConfig(channelID)
	state := m.GetReconnectState(channelID)

	return map[string]interface{}{
		"status":        status,
		"config":        config,
		"reconnect":     state,
		"channel_id":    channelID,
	}
}

func (m *BufferManager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.bufferService.CleanupAllChannels()

	for channelID, state := range m.reconnectStates {
		if time.Since(state.LastReconnect) > 24*time.Hour {
			delete(m.reconnectStates, channelID)
		}
	}
}

func (m *BufferManager) StartCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			m.Cleanup()
		}
	}()
}
