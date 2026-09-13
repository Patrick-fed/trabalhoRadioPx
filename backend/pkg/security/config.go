package security

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
)

type SecurityConfig struct {
	JWTSecret        string
	CORSOrigins      []string
	AllowedMethods   []string
	AllowedHeaders   []string
	MaxBodySize      int64
	EnableHTTPS      bool
	SessionTimeout   int
}

func LoadSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		JWTSecret:      getEnvOrDefault("JWT_SECRET", generateRandomSecret()),
		CORSOrigins:    strings.Split(getEnvOrDefault("CORS_ORIGINS", "*"), ","),
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-User-ID"},
		MaxBodySize:    10 * 1024 * 1024, // 10MB
		EnableHTTPS:    getEnvOrDefault("ENABLE_HTTPS", "false") == "true",
		SessionTimeout: 86400, // 24 hours
	}
}

func (c *SecurityConfig) Validate() bool {
	if c.JWTSecret == "" || c.JWTSecret == "default-secret" {
		return false
	}

	if len(c.JWTSecret) < 32 {
		return false
	}

	return true
}

func (c *SecurityConfig) GetJWTSecret() []byte {
	return []byte(c.JWTSecret)
}

func (c *SecurityConfig) IsOriginAllowed(origin string) bool {
	if len(c.CORSOrigins) == 1 && c.CORSOrigins[0] == "*" {
		return true
	}

	for _, allowed := range c.CORSOrigins {
		if strings.TrimSpace(allowed) == origin {
			return true
		}
	}

	return false
}

func generateRandomSecret() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func SanitizeInput(input string) string {
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, "<script>", "")
	input = strings.ReplaceAll(input, "</script>", "")
	input = strings.ReplaceAll(input, "javascript:", "")
	input = strings.ReplaceAll(input, "onerror=", "")
	input = strings.ReplaceAll(input, "onload=", "")
	return input
}

func IsValidEmail(email string) bool {
	if !strings.Contains(email, "@") {
		return false
	}
	if !strings.Contains(email, ".") {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	if parts[0] == "" || parts[1] == "" {
		return false
	}
	return true
}

func IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			hasUpper = true
		}
		if char >= 'a' && char <= 'z' {
			hasLower = true
		}
		if char >= '0' && char <= '9' {
			hasDigit = true
		}
	}

	return hasUpper && hasLower && hasDigit
}
