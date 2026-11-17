package repository

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Post struct {
	ID        int64    `json:"id"`
	Content   string   `json:"content"`
	Title     string   `json:"title"`
	UserID    int64    `json:"user_id"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}
type PostRepository struct {
	db *sql.DB
}

func (r *PostRepository) Create(ctx context.Context, post *Post) error {
	query := `
		INSERT INTO post (content, title, user_id, tags)
		VSALUES ($1, $2, $3, $4) RETURNING id, created_at, updated-at
	`
	if err := r.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserID,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		post.UpdatedAt,
	); err != nil {
		return err
	}
	return nil
}
