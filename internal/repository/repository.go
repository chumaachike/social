package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("record not found")
	ErrConflict          = errors.New("duplicate key value vioilates unique constraint")
	ErrDuplicateEmail    = errors.New("email has already been used")
	ErrDuplicateUsername = errors.New("username has already been used")
	QueryTimeoutDuration = time.Second * 5
)

type Repository struct {
	Post interface {
		Create(context.Context, *Post) error
		GetByID(context.Context, string) (Post, error)
		DeleteByID(context.Context, string) error
		UpdateByID(context.Context, Post) error
		GetUserFeed(context.Context, string, PaginatedFeedQuery) ([]PostWithMetadata, error)
	}

	User interface {
		Create(context.Context, *sql.Tx, *User) error
		GetUser(context.Context, string) (User, error)
		CreateAndInvite(context.Context, *User, string, time.Duration) error
	}
	Comment interface {
		GetByPostID(context.Context, int64) ([]Comment, error)
	}
	Follower interface {
		Follow(context.Context, string, string) error
		Unfollow(context.Context, string, string) error
	}
}

func NewRepository(db *sql.DB) Repository {
	return Repository{
		Post:     &PostRepository{db},
		User:     &UserRepository{db},
		Comment:  &CommentRepository{db},
		Follower: &FollowerStore{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()

}
