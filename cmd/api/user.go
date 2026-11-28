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

const (
	authUserCtx userKey = "authenticated_user"
	userCtx     userKey = "requested_user"
)

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Produce		json
//	@Param			userID	path		int	true	"User ID"
//	@Success		200		{object}	repository.User
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{string}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		ApiKeyAuth
//	@Router			/users/{userID} [get]
func (app *application) getUser(w http.ResponseWriter, r *http.Request) {
	user, ok := getUserFromContext(r)
	if !ok {
		app.internalServerError(w, r, errors.New("requested user missing from context"))
		return
	}

	if err := app.jsonResponse(
		w, http.StatusOK, user,
	); err != nil {
		app.internalServerError(w, r, err)
	}
}

// FollowUser godoc
//
//	@Summary		Follow a user
//	@Description	Allows an authenticated user to follow another user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			payload	body	FollowUser	true	"User to follow"
//	@Success		204		"No Content"
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		ApiKeyAuth
//	@Router			/users/follow [post]
func (app *application) followUser(w http.ResponseWriter, r *http.Request) {
	user, ok := getAuthenticatedUserFromContext(r)
	if !ok {
		app.unauthorizedResponse(w, r, fmt.Errorf("authentication required"))
		return
	}

	var payload FollowUser

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if user.ID == payload.UserID {
		app.badRequestResponse(w, r, fmt.Errorf("you cannot follow yourself"))
		return
	}

	ctx := r.Context()

	if err := app.repository.Follower.Follow(ctx, user.ID, payload.UserID); err != nil {
		switch {
		case errors.Is(err, repository.ErrConflict):
			app.conflictResponse(w, r, err)

		case errors.Is(err, repository.ErrNotFound):
			app.notFoundResponse(w, r, errors.New("user not found"))
		default:
			app.internalServerError(w, r, err)
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type FollowUser struct {
	// ID of the user to follow or unfollow
	UserID int64 `json:"user_id" validate:"required,gt=0" example:"42"`
}

// UnfollowUser godoc
//
//	@Summary		Unfollow a user
//	@Description	Allows an authenticated user to unfollow another user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			payload	body	FollowUser	true	"User to unfollow"
//	@Success		204		"No Content"
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		ApiKeyAuth
//	@Router			/users/unfollow [post]
func (app *application) unfollowUser(w http.ResponseWriter, r *http.Request) {
	user, ok := getAuthenticatedUserFromContext(r)
	if !ok {
		app.unauthorizedResponse(w, r, fmt.Errorf("authentication required"))
		return
	}

	var payload FollowUser

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	if err := app.repository.Follower.Unfollow(
		ctx,
		user.ID,
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

		id, err := strconv.ParseInt(userID, 10, 64)

		if err != nil || id <= 0 {
			app.badRequestResponse(w, r, errors.New("invalid user ID"))
			return
		}

		ctx := r.Context()

		user, err := app.repository.User.GetUser(ctx, id)

		if err != nil {
			switch {
			case errors.Is(err, repository.ErrNotFound):
				app.notFoundResponse(w, r, errors.New("user not found"))
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

func getAuthenticatedUserFromContext(r *http.Request) (repository.User, bool) {
	user, ok := r.Context().Value(authUserCtx).(repository.User)
	return user, ok
}
