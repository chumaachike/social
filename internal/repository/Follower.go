package repository

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type FOllower struct {
	UserID     string `json:"user_id"`
	FollowerID string `json:"follower_id"`
	CreatedAt  string `json:"created_at"`
}

type FollowerStore struct {
	db *sql.DB
}

func (s *FollowerStore) Follow(ctx context.Context, userID, followerID string) error {
	query := `
        INSERT INTO followers (user_id, follower_id)
        VALUES ($1, $2)
        ON CONFLICT (user_id, follower_id) DO NOTHING
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, userID, followerID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrConflict
		}
	}
	return nil
}

func (s *FollowerStore) Unfollow(ctx context.Context, userID, followerID string) error {
	query := `
        DELETE FROM followers
        WHERE user_id = $1 AND follower_id = $2
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, userID, followerID)
	return err
}
