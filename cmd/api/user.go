package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) getUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	ctx := r.Context()
	user, err := app.repository.User.GetUser(ctx, userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

}
