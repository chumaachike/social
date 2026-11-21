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
	QueryTimeoutDuration = time.Second * 5
)

type Repository struct {
	Post interface {
		Create(context.Context, *Post) error
		GetByID(context.Context, string) (Post, error)
		DeleteByID(context.Context, string) error
		UpdateByID(context.Context, Post) error
	}

	User interface {
		Create(context.Context, *User) error
		GetUser(context.Context, string) (User, error)
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
