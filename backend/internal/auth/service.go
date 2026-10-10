package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type AuthService struct {
	secretKey     []byte
	userService   UserServiceInterface
	tokenExpiry   time.Duration
	refreshExpiry time.Duration
}

type UserServiceInterface interface {
	Authenticate(email, password string) (*User, error)
	GetUser(userID string) (*User, error)
	CreateOAuthUser(email, name, provider, providerID string) (*User, error)
}

type User struct {
	ID       string
	Email    string
	Name     string
	Provider string
}

type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Type     string `json:"type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func NewAuthService(secretKey string, userService UserServiceInterface) *AuthService {
	return &AuthService{
		secretKey:     []byte(secretKey),
		userService:   userService,
		tokenExpiry:   24 * time.Hour,
		refreshExpiry: 7 * 24 * time.Hour,
	}
}

func (s *AuthService) Login(email, password string) (*TokenPair, error) {
	user, err := s.userService.Authenticate(email, password)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokenPair(user)
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateAccessToken verifies an access token and returns the user ID. It
// rejects tokens that are not of the access type (e.g. refresh tokens).
func (s *AuthService) ValidateAccessToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	if claims.Type != TokenTypeAccess {
		return "", ErrInvalidToken
	}

	return claims.UserID, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (*TokenPair, error) {
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	if claims.Type != TokenTypeRefresh {
		return nil, ErrInvalidToken
	}

	user, err := s.userService.GetUser(claims.UserID)
	if err != nil {
		return nil, err
	}

	return s.generateTokenPair(&User{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	})
}

func (s *AuthService) generateTokenPair(user *User) (*TokenPair, error) {
	now := time.Now()
	expiresAt := now.Add(s.tokenExpiry)

	claims := &Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Provider: user.Provider,
		Type:     TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(s.secretKey)
	if err != nil {
		return nil, err
	}

	refreshClaims := &Claims{
		UserID: user.ID,
		Type:   TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(s.secretKey)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *AuthService) GetUserIDFromToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}
