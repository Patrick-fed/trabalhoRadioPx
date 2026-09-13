package voice

import (
	"testing"
	"time"
)

func TestAudioBufferService_AddMessage(t *testing.T) {
	service := NewAudioBufferService(nil)

	msg, err := service.AddMessage("channel1", "user1", []byte("audio-data"))
	if err != nil {
		t.Fatalf("AddMessage failed: %v", err)
	}

	if msg.ChannelID != "channel1" {
		t.Errorf("Expected channel1, got %s", msg.ChannelID)
	}
	if msg.UserID != "user1" {
		t.Errorf("Expected user1, got %s", msg.UserID)
	}
	if msg.Sequence != 1 {
		t.Errorf("Expected sequence 1, got %d", msg.Sequence)
	}
}

func TestAudioBufferService_GetMessages(t *testing.T) {
	service := NewAudioBufferService(nil)

	service.AddMessage("channel1", "user1", []byte("audio1"))
	service.AddMessage("channel1", "user1", []byte("audio2"))
	service.AddMessage("channel1", "user2", []byte("audio3"))

	messages := service.GetMessages("channel1", 0)
	if len(messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(messages))
	}

	messages = service.GetMessages("channel1", 1)
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}
}

func TestAudioBufferService_MarkAsReplayed(t *testing.T) {
	service := NewAudioBufferService(nil)

	msg1, _ := service.AddMessage("channel1", "user1", []byte("audio1"))
	_, _ = service.AddMessage("channel1", "user1", []byte("audio2"))

	service.MarkAsReplayed("channel1", []string{msg1.ID})

	status := service.GetStatus()
	if status.TotalMessages != 2 {
		t.Errorf("Expected 2 total messages, got %d", status.TotalMessages)
	}
}

func TestAudioBufferService_Cleanup(t *testing.T) {
	config := &BufferConfig{
		MaxDuration:    1 * time.Millisecond,
		MaxSizeBytes:   1024,
		CleanupInterval: 1 * time.Minute,
	}
	service := NewAudioBufferService(config)

	service.AddMessage("channel1", "user1", []byte("audio1"))
	time.Sleep(2 * time.Millisecond)

	service.CleanupAllChannels()

	messages := service.GetMessages("channel1", 0)
	if len(messages) != 0 {
		t.Errorf("Expected 0 messages after cleanup, got %d", len(messages))
	}
}

func TestAudioBufferService_GetStatus(t *testing.T) {
	service := NewAudioBufferService(nil)

	service.AddMessage("channel1", "user1", []byte("audio1"))
	service.AddMessage("channel2", "user2", []byte("audio2"))

	status := service.GetStatus()
	if status.TotalMessages != 2 {
		t.Errorf("Expected 2 total messages, got %d", status.TotalMessages)
	}
	if status.ChannelCount != 2 {
		t.Errorf("Expected 2 channels, got %d", status.ChannelCount)
	}
}
