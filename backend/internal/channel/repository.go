package channel

import (
	"database/sql"
	"errors"
	"math"
	"time"
)

// ChannelStore is the persistence boundary of the channel module. Membership
// capacity and uniqueness are enforced inside the store (transactionally in
// PostgreSQL) so concurrent joins cannot oversubscribe a channel.
type ChannelStore interface {
	Create(ch *Channel) error
	GetByID(id string) (*Channel, error)
	AddMember(channelID, userID string) error
	RemoveMember(channelID, userID string) error
	IsMember(channelID, userID string) (bool, error)
	ListMembers(channelID string) ([]string, error)
	ListUserChannels(userID string) ([]*Channel, error)
	ListNearby(latitude, longitude, radiusKm float64) ([]*Channel, error)
	Update(ch *Channel) error
	Delete(id string) error
}

// PGChannelStore persists channels and memberships in PostgreSQL. UserCount is
// never stored: it is computed from user_channels on every read, and
// IsAvailable is derived from it.
type PGChannelStore struct {
	db *sql.DB
}

func NewPGChannelStore(database *sql.DB) *PGChannelStore {
	return &PGChannelStore{db: database}
}

const channelSelect = `
	SELECT c.id, c.name, c.description, c.latitude, c.longitude, c.max_users,
	       (SELECT COUNT(*) FROM user_channels uc WHERE uc.channel_id = c.id) AS user_count,
	       c.created_by, c.created_at, c.updated_at
	FROM channels c`

func (s *PGChannelStore) Create(ch *Channel) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO channels (id, name, description, latitude, longitude, max_users, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		ch.ID, ch.Name, ch.Description, ch.Latitude, ch.Longitude,
		ch.MaxUsers, ch.CreatedBy, ch.CreatedAt, ch.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO user_channels (user_id, channel_id, joined_at)
		VALUES ($1, $2, $3)`,
		ch.CreatedBy, ch.ID, ch.CreatedAt,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *PGChannelStore) GetByID(id string) (*Channel, error) {
	row := s.db.QueryRow(channelSelect+` WHERE c.id = $1`, id)
	ch, err := scanChannel(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrChannelNotFound
	}
	return ch, err
}

func (s *PGChannelStore) AddMember(channelID, userID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var maxUsers int
	err = tx.QueryRow(`SELECT max_users FROM channels WHERE id = $1 FOR UPDATE`, channelID).Scan(&maxUsers)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrChannelNotFound
	}
	if err != nil {
		return err
	}

	result, err := tx.Exec(`
		INSERT INTO user_channels (user_id, channel_id, joined_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, channel_id) DO NOTHING`,
		userID, channelID, time.Now(),
	)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrUserAlreadyInChannel
	}

	var count int
	if err := tx.QueryRow(
		`SELECT COUNT(*) FROM user_channels WHERE channel_id = $1`, channelID,
	).Scan(&count); err != nil {
		return err
	}
	if count > maxUsers {
		return ErrChannelFull
	}

	if _, err := tx.Exec(
		`UPDATE channels SET updated_at = $1 WHERE id = $2`, time.Now(), channelID,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *PGChannelStore) RemoveMember(channelID, userID string) error {
	result, err := s.db.Exec(`
		DELETE FROM user_channels WHERE user_id = $1 AND channel_id = $2`,
		userID, channelID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		var exists bool
		if err := s.db.QueryRow(
			`SELECT EXISTS (SELECT 1 FROM channels WHERE id = $1)`, channelID,
		).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return ErrChannelNotFound
		}
		return ErrUserNotInChannel
	}

	_, err = s.db.Exec(`UPDATE channels SET updated_at = $1 WHERE id = $2`, time.Now(), channelID)
	return err
}

func (s *PGChannelStore) IsMember(channelID, userID string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM user_channels WHERE user_id = $1 AND channel_id = $2
		)`, userID, channelID,
	).Scan(&exists)
	return exists, err
}

func (s *PGChannelStore) ListMembers(channelID string) ([]string, error) {
	var channelExists bool
	if err := s.db.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM channels WHERE id = $1)`, channelID,
	).Scan(&channelExists); err != nil {
		return nil, err
	}
	if !channelExists {
		return nil, ErrChannelNotFound
	}

	rows, err := s.db.Query(`
		SELECT user_id FROM user_channels WHERE channel_id = $1 ORDER BY joined_at`, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]string, 0)
	for rows.Next() {
		var member string
		if err := rows.Scan(&member); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, rows.Err()
}

func (s *PGChannelStore) ListUserChannels(userID string) ([]*Channel, error) {
	rows, err := s.db.Query(channelSelect+`
		JOIN user_channels uc ON uc.channel_id = c.id
		WHERE uc.user_id = $1
		ORDER BY uc.joined_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanChannels(rows)
}

// ListNearby pre-filters with a bounding box (cheap index scan) because the
// exact Haversine filter and the ordering are applied by the service.
func (s *PGChannelStore) ListNearby(latitude, longitude, radiusKm float64) ([]*Channel, error) {
	dLat := radiusKm / 111.0
	cos := math.Cos(latitude * math.Pi / 180)
	if cos < 0.01 {
		cos = 0.01
	}
	dLng := radiusKm / (111.0 * cos)

	rows, err := s.db.Query(channelSelect+`
		WHERE c.latitude BETWEEN $1 AND $2 AND c.longitude BETWEEN $3 AND $4`,
		latitude-dLat, latitude+dLat, longitude-dLng, longitude+dLng,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanChannels(rows)
}

func (s *PGChannelStore) Update(ch *Channel) error {
	result, err := s.db.Exec(`
		UPDATE channels SET name = $1, description = $2, updated_at = $3
		WHERE id = $4`,
		ch.Name, ch.Description, ch.UpdatedAt, ch.ID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrChannelNotFound
	}
	return nil
}

func (s *PGChannelStore) Delete(id string) error {
	result, err := s.db.Exec(`DELETE FROM channels WHERE id = $1`, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrChannelNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanChannel(row rowScanner) (*Channel, error) {
	var ch Channel
	var createdBy sql.NullString
	err := row.Scan(
		&ch.ID, &ch.Name, &ch.Description, &ch.Latitude, &ch.Longitude,
		&ch.MaxUsers, &ch.UserCount, &createdBy, &ch.CreatedAt, &ch.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	ch.CreatedBy = createdBy.String
	ch.IsAvailable = ch.UserCount < ch.MaxUsers
	return &ch, nil
}

func scanChannels(rows *sql.Rows) ([]*Channel, error) {
	channels := make([]*Channel, 0)
	for rows.Next() {
		ch, err := scanChannel(rows)
		if err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}
