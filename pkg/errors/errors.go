package errors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type envelope map[string]any

type ErrorResponse struct {
	logger *zerolog.Logger
}

func NewErrorResponse(logger *zerolog.Logger) ErrorResponse {
	return ErrorResponse{
		logger: logger,
	}
}

func (h *ErrorResponse) LogError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	ctx := r.Context()
	log := log.Ctx(ctx).With().Logger()

	log.Err(err).Msg(fmt.Sprintf("method %s - uri %s", method, uri))
}

func (h *ErrorResponse) ErrorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{"error": message}

	err := writeJSON(w, status, env, nil)
	if err != nil {
		h.LogError(r, err)
		w.WriteHeader(500)
	}
}

func (h *ErrorResponse) ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	h.LogError(r, err)

	message := "the server encountered a problem and could not process your request"
	h.ErrorResponse(w, r, http.StatusInternalServerError, message)
}

func (h *ErrorResponse) NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	h.ErrorResponse(w, r, http.StatusNotFound, message)
}

func (h *ErrorResponse) MethodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	h.ErrorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (h *ErrorResponse) BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	// h.sendReport(err) // This one is just for testing

	h.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (h *ErrorResponse) FailedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	h.ErrorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (h *ErrorResponse) InvalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	message := "invalid or missing authentication token"
	h.ErrorResponse(w, r, http.StatusUnauthorized, message)
}

func (h *ErrorResponse) AuthenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	h.ErrorResponse(w, r, http.StatusUnauthorized, message)
}

func writeJSON(w http.ResponseWriter, status int, data map[string]any, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Append a newline to make it easier to view in terminal applications.
	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}
