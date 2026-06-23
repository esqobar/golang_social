package main

import (
	"errors"
	"net/http"
	"social/internal/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CreateCommentPayload struct {
	Content string `json:"content" validate:"required,max=500"`
}

// createCommentHandler godoc
//
//	@Summary		Create a comment
//	@Description	Add a comment to a post
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		int						true	"Post ID"
//	@Param			payload	body		CreateCommentPayload	true	"Comment payload"
//	@Success		201		{object}	store.Comment			"Comment created"
//	@Failure		400		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/comments/{postId} [post]
func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	// 🔹 Get postID from URL
	postIDParam := chi.URLParam(r, "postID")

	postID, err := strconv.ParseInt(postIDParam, 10, 64)
	if err != nil || postID <= 0 {
		app.badRequestResponse(w, r, errors.New("invalid post id"))
		return
	}

	// 🔹 Parse payload
	var payload CreateCommentPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// 🔹 Validate
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// 🔹 Ensure post exists (important)
	_, err = app.store.Posts.GetById(r.Context(), postID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.notFoundResponse(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	// 🔹 Create model
	comment := &store.Comment{
		PostID:  postID,
		UserID:  1, // replace with authenticated user
		Content: payload.Content,
	}

	// 🔹 Save
	ctx := r.Context()
	if err := app.store.Comments.Create(ctx, comment); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// 🔹 Response
	if err := app.jsonResponse(w, http.StatusCreated, comment); err != nil {
		app.internalServerError(w, r, err)
	}
}
