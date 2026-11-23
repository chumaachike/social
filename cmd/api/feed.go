package main

import (
	"net/http"

	"github.com/chumaachike/social/internal/repository"
)

func (app *application) getUserFeed(w http.ResponseWriter, r *http.Request) {

	fq := repository.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	ctx := r.Context()
	feed, err := app.repository.Post.GetUserFeed(ctx, "1", fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
