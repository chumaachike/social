package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Version   int       `json:"version"`
	Comments  []Comment `json:"comments"`
}

type PostRepository struct {
	db *sql.DB
}

func (r *PostRepository) Create(ctx context.Context, post *Post) error {
	query := `
		INSERT INTO posts (content, title, user_id, tags)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
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
		&post.UpdatedAt,
	); err != nil {
		return err
	}
	return nil
}

func (r *PostRepository) GetByID(ctx context.Context, id string) (Post, error) {
	query := `
		SELECT id, content, title, user_id, tags, version, created_at, updated_at FROM posts WHERE id= $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var post Post
	if err := r.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&post.ID,
		&post.Content,
		&post.Title,
		&post.UserID,
		pq.Array(&post.Tags),
		&post.Version,
		&post.CreatedAt,
		&post.UpdatedAt,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return Post{}, ErrNotFound
		default:
			return Post{}, fmt.Errorf("failed to get post by id: %w", err)
		}
	}
	return post, nil
}

func (r *PostRepository) DeleteByID(ctx context.Context, id string) error {
	query := `
		DELETE FROM posts WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostRepository) UpdateByID(ctx context.Context, post Post) error {
	query := `
		UPDATE posts 
		SET title = $1,
			content = $2, version = version + 1,
			updated_at = NOW()
		WHERE id = $3 AND version = $4
		RETURNING version
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := r.db.QueryRowContext(ctx, query, post.Title, post.Content, post.ID, post.Version).Scan(&post.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}
	return nil
}
