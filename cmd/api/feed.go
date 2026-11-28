package main

import (
	"net/http"

	"github.com/chumaachike/social/internal/repository"
)

// getUserFeed godoc
//
//	@Summary		Get user feed
//	@Description	Retrieves a paginated and sorted user feed.
//	@Tags			Feed
//	@Produce		json
//	@Param			limit	query	int		false	"Number of posts"	default(20) minimum(1) maximum(100)
//	@Param			offset	query	int		false	"Pagination offset"	default(0) minimum(0)
//	@Param			sort	query	string	false	"Sort order"		default(desc) Enums(asc, desc)
//	@Success		200		{array}		repository.Post
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/users/feed [get]
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
