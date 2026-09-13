package voice

import (
	"encoding/json"
	"time"
)

type AudioPacket struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ChannelID string    `json:"channel_id"`
	Audio     []byte    `json:"audio"`
	IsPTT     bool      `json:"is_ptt"`
	Timestamp time.Time `json:"timestamp"`
	Codec     string    `json:"codec"`
}

type AudioMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type AudioPayload struct {
	Audio     []byte `json:"audio"`
	IsPTT     bool   `json:"is_ptt"`
	UserID    string `json:"user_id"`
	ChannelID string `json:"channel_id"`
}

type VoiceStatus struct {
	IsTransmitting bool   `json:"is_transmitting"`
	CurrentSpeaker string `json:"current_speaker,omitempty"`
	ChannelID      string `json:"channel_id"`
}

func NewAudioPacket(userID, channelID string, audio []byte, isPTT bool) *AudioPacket {
	return &AudioPacket{
		ID:        generatePacketID(),
		UserID:    userID,
		ChannelID: channelID,
		Audio:     audio,
		IsPTT:     isPTT,
		Timestamp: time.Now(),
		Codec:     "opus",
	}
}

func (ap *AudioPacket) ToJSON() ([]byte, error) {
	return json.Marshal(ap)
}

func (ap *AudioPacket) ToPayload() AudioPayload {
	return AudioPayload{
		Audio:     ap.Audio,
		IsPTT:     ap.IsPTT,
		UserID:    ap.UserID,
		ChannelID: ap.ChannelID,
	}
}

func generatePacketID() string {
	return time.Now().Format("20060102150405.000000000")
}
