package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"version"`
	Comments  []Comment `json:"comments,omitempty"`
	User      User      `json:"user"`
}

type PostWithMetadata struct {
	Post
	CommentsCount int `json:"comment_count"`
}
type PostRepository struct {
	db *sql.DB
}

func (r *PostRepository) GetUserFeed(
	ctx context.Context,
	userID string,
	fq PaginatedFeedQuery,
) ([]PostWithMetadata, error) {

	sort := strings.ToUpper(fq.Sort)

	if sort != "ASC" && sort != "DESC" {
		sort = "DESC"
	}

	query := `
        SELECT
            p.id,
            p.user_id,
            p.title,
            p.content,
            p.created_at,
            p.version,
            p.tags,
            u.username,
            COUNT(c.id) AS comments_count
        FROM posts p
        LEFT JOIN comments c ON c.post_id = p.id
        JOIN users u ON p.user_id = u.id
        WHERE (
            p.user_id = $1
            OR EXISTS (
                SELECT 1
                FROM followers f
                WHERE f.user_id = p.user_id
                AND f.follower_id = $1
            )
        )
        AND (
            p.title ILIKE '%' || $4 || '%'
            OR p.content ILIKE '%' || $4 || '%'
        )
        AND (
			$5::varchar[] IS NULL
			OR $5::varchar[] = '{}'
			OR p.tags @> $5::varchar[]
		)
        GROUP BY p.id, u.username
        ORDER BY p.created_at ` + sort + `, p.id ` + sort + `
        LIMIT $2 OFFSET $3
    `

	ctx, cancel := context.WithTimeout(
		ctx,
		QueryTimeoutDuration,
	)
	defer cancel()

	rows, err := r.db.QueryContext(
		ctx,
		query,
		userID,
		fq.Limit,
		fq.Offset,
		fq.Search,
		pq.Array(fq.Tags),
	)
	if err != nil {
		return nil, fmt.Errorf("query user feed: %w", err)
	}
	defer rows.Close()

	feed := make([]PostWithMetadata, 0)

	for rows.Next() {
		var post PostWithMetadata

		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.Version,
			pq.Array(&post.Tags),
			&post.User.Username,
			&post.CommentsCount,
		); err != nil {
			return nil, fmt.Errorf("scan feed post: %w", err)
		}

		feed = append(feed, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user feed: %w", err)
	}

	return feed, nil
}

func (r *PostRepository) Create(ctx context.Context, post *Post) error {
	query := `
		INSERT INTO posts (content, title, user_id, tags)
		VALUES ($1, $2, $3, $4)
		RETURNING id, version, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := r.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserID,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.Version,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create post: %w", err)
	}

	return nil
}

func (r *PostRepository) GetByID(ctx context.Context, id int64) (Post, error) {
	query := `
		SELECT id, content, title, user_id, tags, version, created_at, updated_at
		FROM posts
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var post Post
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Content,
		&post.Title,
		&post.UserID,
		pq.Array(&post.Tags),
		&post.Version,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Post{}, ErrNotFound
		}
		return Post{}, fmt.Errorf("get post by id: %w", err)
	}

	return post, nil
}

func (r *PostRepository) UpdateByID(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts
		SET title = $1,
		    content = $2,
		    version = version + 1,
		    updated_at = NOW()
		WHERE id = $3 AND version = $4
		RETURNING version, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := r.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.ID,
		post.Version,
	).Scan(
		&post.Version,
		&post.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return r.resolveUpdateNoRows(ctx, post.ID)
		}
		return fmt.Errorf("update post: %w", err)
	}

	return nil
}

func (r *PostRepository) DeleteByID(ctx context.Context, id int64) error {
	query := `
		DELETE FROM posts
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete post rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostRepository) resolveUpdateNoRows(ctx context.Context, id int64) error {
	var exists bool
	err := r.db.QueryRowContext(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)`,
		id,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check post existence on update: %w", err)
	}

	if !exists {
		return ErrNotFound
	}

	return ErrConflict
}
