package channel

import (
	"time"
)

type Channel struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Latitude    float64   `json:"latitude" db:"latitude"`
	Longitude   float64   `json:"longitude" db:"longitude"`
	MaxUsers    int       `json:"max_users" db:"max_users"`
	UserCount   int       `json:"user_count" db:"user_count"`
	IsAvailable bool      `json:"is_available" db:"is_available"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateChannelRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Latitude    float64 `json:"latitude" binding:"required"`
	Longitude   float64 `json:"longitude" binding:"required"`
}

type UpdateChannelRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ChannelFilter struct {
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	RadiusKm   float64 `json:"radius_km"`
	MaxResults int     `json:"max_results"`
}
