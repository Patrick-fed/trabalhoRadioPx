package voice

import (
	"time"

	ws "github.com/radiopx/backend/pkg/websocket"
)

// bufferAdapter exposes *AudioBufferService through the ws.AudioBuffer
// interface; pkg/websocket cannot import internal/voice (import cycle).
type bufferAdapter struct {
	svc *AudioBufferService
}

func NewBufferAdapter(svc *AudioBufferService) ws.AudioBuffer {
	return &bufferAdapter{svc: svc}
}

func (a *bufferAdapter) AddMessage(channelID, userID string, audioData []byte) (int64, time.Time, error) {
	msg, err := a.svc.AddMessage(channelID, userID, audioData)
	if err != nil {
		return 0, time.Time{}, err
	}
	return msg.Sequence, msg.Timestamp, nil
}

func (a *bufferAdapter) GetMessages(channelID string, sinceSequence int64) []ws.BufferedAudio {
	msgs := a.svc.GetMessages(channelID, sinceSequence)
	result := make([]ws.BufferedAudio, 0, len(msgs))
	for _, m := range msgs {
		result = append(result, ws.BufferedAudio{
			Audio:     m.AudioData,
			UserID:    m.UserID,
			Sequence:  m.Sequence,
			Timestamp: m.Timestamp,
		})
	}
	return result
}
