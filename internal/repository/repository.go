package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("record not found")
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
}

func NewRepository(db *sql.DB) Repository {
	return Repository{
		Post:    &PostRepository{db},
		User:    &UserRepository{db},
		Comment: &CommentRepository{db},
	}
}
