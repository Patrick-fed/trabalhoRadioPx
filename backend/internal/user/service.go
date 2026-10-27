package user

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
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

// UserService holds user business rules (validation, normalization, bcrypt).
// Persistence goes through UserStore: PostgreSQL in production, in-memory for
// the unit tests. Concurrency control lives in the store.
type UserService struct {
	store UserStore
}

func NewUserService(store UserStore) *UserService {
	return &UserService{store: store}
}

func (s *UserService) CreateUser(req *CreateUserRequest) (*User, error) {
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

	if err := s.store.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUser(userID string) (*User, error) {
	return s.store.GetByID(userID)
}

func (s *UserService) GetUserByEmail(email string) (*User, error) {
	return s.store.GetByEmail(normalizeEmail(email))
}

func (s *UserService) UpdateUser(userID string, req *UpdateUserRequest) error {
	user, err := s.store.GetByID(userID)
	if err != nil {
		return err
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

	return s.store.Update(user)
}

func (s *UserService) DeleteUser(userID string) error {
	return s.store.Delete(userID)
}

func (s *UserService) Authenticate(email, password string) (*User, error) {
	foundUser, err := s.store.GetByEmail(normalizeEmail(email))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return foundUser, nil
}

func (s *UserService) ChangePassword(userID, oldPassword, newPassword string) error {
	user, err := s.store.GetByID(userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	if newPassword == "" {
		return ErrPasswordRequired
	}
	if len(newPassword) < minPasswordLength {
		return ErrPasswordTooShort
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	return s.store.Update(user)
}

func (s *UserService) CreateOAuthUser(email, name, provider, providerID string) (*User, error) {
	email = normalizeEmail(email)
	if existing, err := s.store.GetByEmailAndProvider(email, provider); err == nil {
		return existing, nil
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

	if err := s.store.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) ListUsers() ([]*User, error) {
	return s.store.List()
}

func generateUserID() string {
	return uuid.NewString()
}
