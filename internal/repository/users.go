package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
}
type password struct {
	text *string
	hash []byte
}

type UserRepository struct {
	db *sql.DB
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.text = &text
	p.hash = hash

	return nil
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users(username, password, email) VALUES ($1, $2, $3) RETURNING id, created_at
	`
	if err := r.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Password,
		user.Email,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	); err != nil {
		return err
	}
	return nil
}

func (r UserRepository) GetUser(ctx context.Context, userID string) (User, error) {
	query := `
		SELECT id, username, email FROM users WHERE id = $1
	`
	var user User
	if err := r.db.QueryRowContext(
		ctx,
		query,
		userID,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return User{}, ErrNotFound
		default:
			return User{}, fmt.Errorf("failed to get User by id: %w", err)
		}
	}
	return user, nil
}

func (r *UserRepository) CreateAndInvite(ctx context.Context, user *User, token string) error {
	return nil
}
