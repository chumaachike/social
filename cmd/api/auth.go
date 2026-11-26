package main

import (
	"net/http"

	"github.com/chumaachike/social/internal/repository"
)

type RegisterUserPayload struct {
	Email    string `json:"email" validate:"required,max=100"`
	Username string `json:"username" validate:"required,email=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

// registerUserHandler godoc
// @Summary Regiseter a user
// @Description Registers a user
// @Tags authentication
// @Accept json
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(w, r, payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &repository.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}
	ctx := r.Context()

	if err := app.repository.User.CreateAndInvite(ctx, user, "token-121"); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, user); err != nil {
		app.internalServerError(w, r, err)
	}

}
