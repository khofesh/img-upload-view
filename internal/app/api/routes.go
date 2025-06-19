package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/khofesh/img-upload-view/internal/app/api/handlers"
	"github.com/khofesh/img-upload-view/internal/config"
	"github.com/khofesh/img-upload-view/internal/data"
	middlewares "github.com/khofesh/img-upload-view/internal/middleware"
)

func routes(app *config.Application) http.Handler {
	router := httprouter.New()

	mw := middlewares.New(
		middlewares.WithTrustedOrigins[data.Models](app.Config.TrustedOrigins),
		middlewares.WithErrorResponse[data.Models](app.ErrorResponse),
	)

	router.HandlerFunc(http.MethodPost, "/upload", handlers.UploadImage(app))
	router.HandlerFunc(http.MethodGet, "/image", handlers.GetImage(app))

	return mw.RecoverPanic(router)
}
