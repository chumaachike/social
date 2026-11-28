package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
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

func (r *UserRepository) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		INSERT INTO users(username, password, email) VALUES ($1, $2, $3) RETURNING id, created_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Password.hash,
		user.Email,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}
	return nil
}

func (r UserRepository) GetUser(ctx context.Context, userID int64) (User, error) {
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

func (r *UserRepository) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	return withTx(r.db, ctx, func(tx *sql.Tx) error {
		if err := r.Create(ctx, tx, user); err != nil {
			return err
		}

		if err := r.createUserInvitation(ctx, tx, token, invitationExp, user.ID); err != nil {
			return err
		}
		return nil
	})
}

func (r *UserRepository) Activate(ctx context.Context, token string) error {
	return withTx(r.db, ctx, func(tx *sql.Tx) error {
		//1. find the user that this token belogs to
		user, err := r.findUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}
		// 2. Update the user
		user.IsActive = true
		if err := r.update(ctx, tx, user); err != nil {
			return err
		}

		// 3. Delete user invitations
		if err := r.deleteUserInvitations(ctx, tx, user.ID); err != nil {
			return err
		}
		return nil

	})

}

func (r *UserRepository) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, invitationExp time.Duration, userID int64) error {
	query := `INSERT INTO user_invitations (token, user_id, expiry) VALUES ($1, $2, $3)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, userID, time.Now().Add(invitationExp))

	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) findUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.created_at, u.is_active
		FROM users u
		JOIN user_invitations ui ON u.id = ui.user_id
		WHERE ui.token = $1 AND ui.expiry > $2
	`
	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
	)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return user, nil
}

func (r *UserRepository) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `UPDATE users SET username = $1, email = $2, is_active = $3 WHERE id = $4`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)
	if err != nil {
		return nil
	}
	return nil
}

func (r *UserRepository) deleteUserInvitations(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM user_invitations WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}
	return nil
}
