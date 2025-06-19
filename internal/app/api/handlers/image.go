package handlers

import (
	"net/http"

	"github.com/khofesh/img-upload-view/internal/config"
	"github.com/khofesh/img-upload-view/internal/reqres"
)

type envelope map[string]any

func UploadImage(app *config.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := reqres.WriteJSON(w, http.StatusOK, envelope{"message": "ok"}, nil)

		if err != nil {
			app.ErrorResponse.ServerErrorResponse(w, r, err)
		}
	}
}

func GetImage(app *config.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := reqres.WriteJSON(w, http.StatusOK, envelope{"message": "ok"}, nil)

		if err != nil {
			app.ErrorResponse.ServerErrorResponse(w, r, err)
		}
	}
}
