package main

import (
	"net/http"
)

func (app *application) errorResponse(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	message string,
	err error,
) {
	// Centralized structured logging
	if err != nil {
		switch {
		case status >= 500:
			app.logger.Errorw("server error",
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"error", err,
			)
		case status >= 400:
			app.logger.Warnw("client error",
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"error", err,
			)
		default:
			app.logger.Infow("response",
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
			)
		}
	}

	// Consistent JSON error response
	writeJSONError(w, status, message)
}

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(
		w,
		r,
		http.StatusInternalServerError,
		"The server encountered a problem and could not process your request",
		err,
	)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(
		w,
		r,
		http.StatusBadRequest,
		err.Error(),
		err,
	)
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(
		w,
		r,
		http.StatusNotFound,
		"The requested resource could not be found",
		err,
	)
}

func (app *application) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(
		w,
		r,
		http.StatusConflict,
		"The resource already exists",
		err,
	)
}

func (app *application) unAuthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(
		w,
		r,
		http.StatusUnauthorized,
		"Unauthorized error",
		err,
	)
}

func (app *application) unAuthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted", charset="UTF-8"`)

	app.errorResponse(
		w,
		r,
		http.StatusUnauthorized,
		"Unauthorized basic error",
		err,
	)
}

func (app *application) forbiddenResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(
		w,
		r,
		http.StatusForbidden,
		"Forbidden error",
		err,
	)
}

func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	w.Header().Set("Retry-After", retryAfter)

	app.errorResponse(
		w,
		r,
		http.StatusTooManyRequests,
		"rate limit exceeded, retry after: "+retryAfter,
		nil,
	)
}

//func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
//	app.logger.Warnw("rate limit exceeded", "method", r.Method, "path", r.URL.Path)
//
//	w.Header().Set("Retry-After", retryAfter)
//
//	writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded, retry after: "+retryAfter)
//}

//func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
//
//	app.logger.Errorw("Internal server error", "method", r.Method, "path", r.URL.Path, "error", err)
//	writeJSONError(w, http.StatusInternalServerError, "internal server encountered a problem")
//}
//
//func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
//	app.logger.Warnf("Bad request", "method", r.Method, "path", r.URL.Path, "error", err)
//	writeJSONError(w, http.StatusBadRequest, err.Error())
//}
//
//func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
//	app.logger.Warnf("Not found error", "method", r.Method, "path", r.URL.Path, "error", err)
//	writeJSONError(w, http.StatusNotFound, "not found")
//}
//
//func (app *application) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
//	app.logger.Errorf("The resource already exists", "method", r.Method, "path", r.URL.Path, "error", err)
//	writeJSONError(w, http.StatusNotFound, "not found")
//}
