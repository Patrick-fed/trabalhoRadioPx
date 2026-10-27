package user

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// UserStore is the persistence boundary of the user module. The service holds
// business rules (hashing, validation); the store only reads and writes rows.
type UserStore interface {
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByEmailAndProvider(email, provider string) (*User, error)
	Update(user *User) error
	Delete(id string) error
	List() ([]*User, error)
}

const userColumns = `id, email, password_hash, name, avatar, provider, provider_id, is_active, created_at, updated_at`

// PGUserStore persists users in PostgreSQL.
type PGUserStore struct {
	db *sql.DB
}

func NewPGUserStore(database *sql.DB) *PGUserStore {
	return &PGUserStore{db: database}
}

func (s *PGUserStore) Create(user *User) error {
	_, err := s.db.Exec(`
		INSERT INTO users (id, email, password_hash, name, avatar, provider, provider_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		user.ID, user.Email, user.Password, user.Name, user.Avatar,
		user.Provider, user.ProviderID, user.IsActive, user.CreatedAt, user.UpdatedAt,
	)
	if isUniqueViolation(err) {
		return ErrUserAlreadyExists
	}
	return err
}

func (s *PGUserStore) GetByID(id string) (*User, error) {
	return s.queryOne(`SELECT `+userColumns+` FROM users WHERE id = $1`, id)
}

func (s *PGUserStore) GetByEmail(email string) (*User, error) {
	return s.queryOne(`SELECT `+userColumns+` FROM users WHERE email = $1`, email)
}

func (s *PGUserStore) GetByEmailAndProvider(email, provider string) (*User, error) {
	return s.queryOne(
		`SELECT `+userColumns+` FROM users WHERE email = $1 AND provider = $2`,
		email, provider,
	)
}

func (s *PGUserStore) Update(user *User) error {
	result, err := s.db.Exec(`
		UPDATE users
		SET name = $1, email = $2, avatar = $3, provider = $4, provider_id = $5,
		    is_active = $6, password_hash = $7, updated_at = $8
		WHERE id = $9`,
		user.Name, user.Email, user.Avatar, user.Provider, user.ProviderID,
		user.IsActive, user.Password, user.UpdatedAt, user.ID,
	)
	if isUniqueViolation(err) {
		return ErrUserAlreadyExists
	}
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (s *PGUserStore) Delete(id string) error {
	result, err := s.db.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (s *PGUserStore) List() ([]*User, error) {
	rows, err := s.db.Query(`SELECT ` + userColumns + ` FROM users ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s *PGUserStore) queryOne(query string, args ...interface{}) (*User, error) {
	row := s.db.QueryRow(query, args...)
	user, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return user, err
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanUser(row rowScanner) (*User, error) {
	var user User
	err := row.Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.Avatar,
		&user.Provider, &user.ProviderID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
