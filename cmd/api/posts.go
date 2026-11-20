package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/chumaachike/social/internal/repository"
	"github.com/go-chi/chi/v5"
)

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

type UpdatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

type contextKey string

const postCtxKey contextKey = "post"

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

func (app *application) getPost(w http.ResponseWriter, r *http.Request) {
	post, ok := getPostFromCtx(r)
	if !ok {
		app.notFoundResponse(w, r, errors.New("post not found"))
	}

	// Get comments linking post
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

func (app *application) DeletePost(w http.ResponseWriter, r *http.Request) {
	postId := chi.URLParam(r, "postID")
	ctx := r.Context()
	if err := app.repository.Post.DeleteByID(ctx, postId); err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
	if err := app.jsonResponse(w, http.StatusOK, "Post deleted"); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) UpdatePost(w http.ResponseWriter, r *http.Request) {
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

	if err := app.repository.Post.UpdateByID(r.Context(), post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) postContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "postID")

		ctx := r.Context()

		post, err := app.repository.Post.GetByID(ctx, idParam)
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
