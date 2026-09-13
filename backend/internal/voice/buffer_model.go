package voice

import (
	"time"
)

type MessageBuffer struct {
	ID          string    `json:"id" db:"id"`
	ChannelID   string    `json:"channel_id" db:"channel_id"`
	UserID      string    `json:"user_id" db:"user_id"`
	AudioData   []byte    `json:"audio_data" db:"audio_data"`
	Sequence    int64     `json:"sequence" db:"sequence"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
	IsReplayed  bool      `json:"is_replayed" db:"is_replayed"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type BufferConfig struct {
	MaxDuration    time.Duration `json:"max_duration"`
	MaxSizeBytes   int64         `json:"max_size_bytes"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

func DefaultBufferConfig() *BufferConfig {
	return &BufferConfig{
		MaxDuration:    30 * time.Second,
		MaxSizeBytes:   10 * 1024 * 1024, // 10MB
		CleanupInterval: 5 * time.Minute,
	}
}

type BufferStatus struct {
	TotalMessages   int           `json:"total_messages"`
	TotalSizeBytes  int64         `json:"total_size_bytes"`
	OldestMessage   *time.Time    `json:"oldest_message"`
	NewestMessage   *time.Time    `json:"newest_message"`
	ChannelCount    int           `json:"channel_count"`
}
