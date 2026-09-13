package location

import (
	"time"
)

type Location struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Latitude  float64   `json:"latitude" db:"latitude"`
	Longitude float64   `json:"longitude" db:"longitude"`
	Accuracy  float64   `json:"accuracy" db:"accuracy"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type LocationUpdate struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Accuracy  float64 `json:"accuracy"`
}

type NearbyChannel struct {
	ChannelID  string  `json:"channel_id"`
	Name       string  `json:"name"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Distance   float64 `json:"distance"`
	UserCount  int     `json:"user_count"`
	MaxUsers   int     `json:"max_users"`
	IsAvailable bool   `json:"is_available"`
}

type ProximityFilter struct {
	RadiusKm   float64 `json:"radius_km"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	MaxResults int     `json:"max_results"`
}
