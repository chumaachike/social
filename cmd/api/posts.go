package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/chumaachike/social/internal/repository"
	"github.com/go-chi/chi/v5"
)

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

type UpdatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty,min=1,max=100"`
	Content *string `json:"content" validate:"omitempty,min=1,max=1000"`
}

func (p UpdatePostPayload) isEmpty() bool {
	return p.Title == nil && p.Content == nil
}

type contextKey string

const postCtxKey contextKey = "post"

// createPost godoc
//
//	@Summary		Create a post
//	@Description	Creates a new social media post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreatePostPayload	true	"Post payload"
//	@Success		201		{object}	repository.Post
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/posts [post]
func (app *application) createPost(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	post := &repository.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  1,
	}
	ctx := r.Context()

	if err := app.repository.Post.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// getPost godoc
//
//	@Summary		Get a post
//	@Description	Fetches a post by ID along with its comments
//	@Tags			posts
//	@Produce		json
//	@Param			postID	path		int	true	"Post ID"
//	@Success		200		{object}	repository.Post
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/posts/{postID} [get]
func (app *application) getPost(w http.ResponseWriter, r *http.Request) {
	post, ok := getPostFromCtx(r)
	if !ok {
		app.notFoundResponse(w, r, errors.New("post not found"))
		return
	}

	comments, err := app.repository.Comment.GetByPostID(r.Context(), post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	post.Comments = comments

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// deletePost godoc
//
//	@Summary		Delete a post
//	@Description	Deletes a post by ID
//	@Tags			posts
//	@Param			postID	path	int	true	"Post ID"
//	@Success		204		"Post deleted"
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/posts/{postID} [delete]
func (app *application) deletePost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "postID"), 10, 64)
	if err != nil || id <= 0 {
		app.badRequestResponse(w, r, errors.New("invalid post ID"))
		return
	}

	ctx := r.Context()
	if err := app.repository.Post.DeleteByID(ctx, id); err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// updatePost godoc
//
//	@Summary		Update a post
//	@Description	Partially updates a post using optimistic concurrency control
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		int					true	"Post ID"
//	@Param			payload	body		UpdatePostPayload	true	"Update post payload"
//	@Success		200		{object}	repository.Post
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/posts/{postID} [patch]
func (app *application) updatePost(w http.ResponseWriter, r *http.Request) {
	post, ok := getPostFromCtx(r)
	if !ok {
		app.notFoundResponse(w, r, errors.New("post not found"))
		return
	}

	var payload UpdatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if payload.isEmpty() {
		app.badRequestResponse(w, r, errors.New("request body must contain at least one field to update"))
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}

	if payload.Content != nil {
		post.Content = *payload.Content
	}

	err := app.repository.Post.UpdateByID(r.Context(), &post)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrConflict):
			app.conflictResponse(w, r, err)
		case errors.Is(err, repository.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) postContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "postID"), 10, 64)
		if err != nil || id <= 0 {
			app.badRequestResponse(w, r, errors.New("invalid post ID"))
			return
		}

		ctx := r.Context()

		post, err := app.repository.Post.GetByID(ctx, id)
		if err != nil {
			switch {
			case errors.Is(err, repository.ErrNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, postCtxKey, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) (repository.Post, bool) {
	post, ok := r.Context().Value(postCtxKey).(repository.Post)
	return post, ok
}
