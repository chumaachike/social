package repository

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrNotFound = errors.New("record not found")
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
