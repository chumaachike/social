package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Follower struct {
	UserID     int64     `json:"user_id"`
	FollowerID int64     `json:"follower_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type FollowerStore struct {
	db *sql.DB
}

func (s *FollowerStore) Follow(ctx context.Context, followerID, targetUserID int64) error {
	query := `
        INSERT INTO followers (user_id, follower_id)  VALUES ($1, $2)
    `
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, targetUserID, followerID)

	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return ErrConflict

			case "foreign_key_violation":
				return ErrNotFound
			}
		}

		return fmt.Errorf(
			"failed to follow user: %w", err,
		)
	}

	return nil
}

func (s *FollowerStore) Unfollow(ctx context.Context, followerID, targetUserID int64) error {
	query := `
        DELETE FROM followers WHERE user_id = $1 AND follower_id = $2
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, targetUserID, followerID)

	if err != nil {
		return fmt.Errorf(
			"failed to unfollow user: %w", err,
		)
	}

	return nil
}
