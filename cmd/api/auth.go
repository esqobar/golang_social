package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"social/internal/mailer"
	"social/internal/store"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}

// registerUserHandler godoc
//
//	@Summary		Registers a user
//	@Description	Registers a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	UserWithToken		"User registered"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/signup [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.notFoundResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
		Role: store.Role{
			Name: "user",
		},
	}

	// Hashing the password
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()
	plainToken := uuid.New().String()

	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	// Store user
	err := app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.exp)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail, store.ErrDuplicateUsername:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	userWithToken := UserWithToken{
		User:  user,
		Token: plainToken,
	}

	activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken)
	isProdEnv := app.config.env == "production"
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}

	// Sending invite mail
	status, err := app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	if err != nil {
		app.logger.Errorw("error sending welcome email", "error", err)

		//rolling back creation if email fails (SAGA Pattern)
		if err := app.store.Users.Delete(ctx, user.ID); err != nil {
			app.logger.Errorw("error deleting user", "error", err)
		}

		app.internalServerError(w, r, err)
		return
	}

	app.logger.Infow("Email sent", "status code", status)

	if err := app.jsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

//func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
//	var payload RegisterUserPayload
//	if err := readJSON(w, r, &payload); err != nil {
//		app.notFoundResponse(w, r, err)
//		return
//	}
//
//	if err := Validate.Struct(payload); err != nil {
//		app.badRequestResponse(w, r, err)
//		return
//	}
//
//	user := &store.User{
//		Username: payload.Username,
//		Email:    payload.Email,
//	}
//
//	//hashing the password
//	if err := user.Password.Set(payload.Password); err != nil {
//		app.internalServerError(w, r, err)
//		return
//	}
//
//	ctx := r.Context()
//
//	plainToken := uuid.New().String()
//
//	//store
//	hash := sha256.Sum256([]byte(plainToken))
//	hashToken := hex.EncodeToString(hash[:])
//
//	//store user
//	err := app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.exp)
//	if err != nil {
//		switch err {
//		case store.ErrDuplicateEmail:
//			app.badRequestResponse(w, r, err)
//		case store.ErrDuplicateUsername:
//			app.badRequestResponse(w, r, err)
//		default:
//			app.internalServerError(w, r, err)
//		}
//		return
//	}
//
//	userWithToken := UserWithToken{
//		User:  user,
//		Token: plainToken,
//	}
//
//	activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken)
//
//	isProdEnv := app.config.env == "production"
//	vars := struct {
//		Username      string
//		ActivationURL string
//	}{
//		Username:      user.Username,
//		ActivationURL: activationURL,
//	}
//	//sending invite mail
//	var status int
//	status, err = app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
//	if err != nil {
//		app.logger.Errorw("error sending welcome email", "error", err)
//
//		//rolling back creation if email fails (SAGA Pattern)
//		if err := app.store.Users.Delete(ctx, user.ID); err != nil {
//			app.logger.Errorw("error deleting user", "error", err)
//		}
//
//		app.internalServerError(w, r, err)
//		return
//	}
//
//	app.logger.Infow("Email sent", "status code", status)
//
//	if err := app.jsonResponse(w, http.StatusCreated, userWithToken); err != nil {
//		app.internalServerError(w, r, err)
//		return
//	}
//}

// activateUserHandler godoc
//
//	@Summary		Activates/Registers a user
//	@Description	Activates/Registers a user by invitation token
//	@Tags			users
//	@Produce		json
//	@Param			token	path		string	true	"Invitation token"
//	@Success		201		{object}	string	"User activated"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	err := app.store.Users.Activate(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

type CreateUserTokenPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

// createTokenHandler godoc
//
//	@Summary		Creates a token
//	@Description	Creates a token for a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateUserTokenPayload	true	"User credentials"
//	@Success		200		{object}	string					"Token"
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/token [post]
func (app *application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateUserTokenPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.notFoundResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	//fetching the user(check if the user exists) from the payload
	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.unAuthorizedErrorResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
	//generating the token -> add claims
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.config.auth.token.exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.auth.token.iss,
		"aud": app.config.auth.token.iss,
	}

	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	//sending it yo the client
	if err := app.jsonResponse(w, http.StatusCreated, token); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
