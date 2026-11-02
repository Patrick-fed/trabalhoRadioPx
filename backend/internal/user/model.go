package user

import (
	"time"
)

type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	Name      string    `json:"name" db:"name"`
	Avatar    string    `json:"avatar" db:"avatar"`
	Provider  string    `json:"provider" db:"provider"`
	ProviderID string   `json:"provider_id" db:"provider_id"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	IsAdmin   bool      `json:"is_admin" db:"is_admin"`
	// Last-known position, set passively from channel API requests. Nil until
	// the user's app first sends a location; never a history (RF09/DEC-018).
	LastLat        *float64   `json:"last_lat" db:"last_lat"`
	LastLng        *float64   `json:"last_lng" db:"last_lng"`
	LastLocationAt *time.Time `json:"last_location_at" db:"last_location_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

type UpdateUserRequest struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	User      *User  `json:"user"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
}
