package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/chumaachike/social/internal/repository"
	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtx userKey = "user"

func (app *application) getUser(w http.ResponseWriter, r *http.Request) {
	user, ok := getUserFromContext(r)
	if !ok {
		app.badRequestResponse(w, r, fmt.Errorf("no userID provided"))
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) followUser(w http.ResponseWriter, r *http.Request) {
	user, ok := getUserFromContext(r)
	if !ok {
		app.badRequestResponse(w, r, fmt.Errorf("authentication required"))
		return
	}

	var payload FollowUser

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if payload.UserID == "" {
		app.badRequestResponse(w, r, fmt.Errorf("user_id is required"))
		return
	}

	followerID := strconv.FormatInt(int64(user.ID), 10)

	if followerID == payload.UserID {
		app.badRequestResponse(w, r, fmt.Errorf("you cannot follow yourself"))
		return
	}

	ctx := r.Context()

	if err := app.repository.Follower.Follow(
		ctx,
		followerID,
		payload.UserID,
	); err != nil {
		switch err {
		case repository.ErrConflict:
			app.conflictResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

type FollowUser struct {
	UserID string `json:"user_id"`
}

func (app *application) unfollowUser(w http.ResponseWriter, r *http.Request) {
	user, ok := getUserFromContext(r)
	if !ok {
		app.badRequestResponse(w, r, fmt.Errorf("authentication required"))
		return
	}

	var payload FollowUser

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if payload.UserID == "" {
		app.badRequestResponse(w, r, fmt.Errorf("user_id is required"))
		return
	}

	followerID := strconv.FormatInt(int64(user.ID), 10)

	ctx := r.Context()

	if err := app.repository.Follower.Unfollow(
		ctx,
		followerID,
		payload.UserID,
	); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")

		if userID == "" {
			app.badRequestResponse(w, r, fmt.Errorf("missing user ID"))
			return
		}

		ctx := r.Context()

		user, err := app.repository.User.GetUser(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, repository.ErrNotFound):
				http.Error(w, "user not found", http.StatusNotFound)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, userCtx, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromContext(r *http.Request) (repository.User, bool) {
	user, ok := r.Context().Value(userCtx).(repository.User)
	return user, ok
}
