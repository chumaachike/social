package repository

import (
	"context"
	"database/sql"
)

type PostRepository struct {
	db *sql.DB
}

func (r *PostRepository) Create(ctx context.Context) error {
	return nil
}
