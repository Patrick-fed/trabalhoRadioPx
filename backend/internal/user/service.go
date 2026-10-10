package user

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("usuário não encontrado")
	ErrUserAlreadyExists  = errors.New("este e-mail já está cadastrado")
	ErrInvalidCredentials = errors.New("e-mail ou senha inválidos")
	ErrEmailRequired      = errors.New("o e-mail é obrigatório")
	ErrPasswordRequired   = errors.New("a senha é obrigatória")
	ErrInvalidEmail       = errors.New("e-mail inválido")
	ErrPasswordTooShort   = errors.New("a senha deve ter pelo menos 8 caracteres")
)

const minPasswordLength = 8

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

type UserService struct {
	mu    sync.RWMutex
	users map[string]*User
}

func NewUserService() *UserService {
	return &UserService{
		users: make(map[string]*User),
	}
}

func (s *UserService) CreateUser(req *CreateUserRequest) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	req.Email = normalizeEmail(req.Email)
	if req.Email == "" {
		return nil, ErrEmailRequired
	}
	if !validEmail(req.Email) {
		return nil, ErrInvalidEmail
	}
	if req.Password == "" {
		return nil, ErrPasswordRequired
	}
	if len(req.Password) < minPasswordLength {
		return nil, ErrPasswordTooShort
	}

	for _, user := range s.users {
		if user.Email == req.Email {
			return nil, ErrUserAlreadyExists
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &User{
		ID:        generateUserID(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		Name:      req.Name,
		Provider:  "local",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.users[user.ID] = user

	return user, nil
}

func (s *UserService) GetUser(userID string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[userID]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (s *UserService) GetUserByEmail(email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	email = normalizeEmail(email)
	for _, user := range s.users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, ErrUserNotFound
}

func (s *UserService) UpdateUser(userID string, req *UpdateUserRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userID]
	if !exists {
		return ErrUserNotFound
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = normalizeEmail(req.Email)
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	user.UpdatedAt = time.Now()

	return nil
}

func (s *UserService) DeleteUser(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.users[userID]
	if !exists {
		return ErrUserNotFound
	}

	delete(s.users, userID)

	return nil
}

func (s *UserService) Authenticate(email, password string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	email = normalizeEmail(email)
	var foundUser *User
	for _, user := range s.users {
		if user.Email == email {
			foundUser = user
			break
		}
	}

	if foundUser == nil {
		return nil, ErrInvalidCredentials
	}

	err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return foundUser, nil
}

func (s *UserService) ChangePassword(userID, oldPassword, newPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userID]
	if !exists {
		return ErrUserNotFound
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	return nil
}

func (s *UserService) CreateOAuthUser(email, name, provider, providerID string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	email = normalizeEmail(email)
	for _, user := range s.users {
		if user.Email == email && user.Provider == provider {
			return user, nil
		}
	}

	now := time.Now()
	user := &User{
		ID:         generateUserID(),
		Email:      email,
		Name:       name,
		Provider:   provider,
		ProviderID: providerID,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	s.users[user.ID] = user

	return user, nil
}

func (s *UserService) ListUsers() []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	return users
}

func generateUserID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
