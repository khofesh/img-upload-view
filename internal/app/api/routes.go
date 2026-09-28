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

	router.HandlerFunc(http.MethodPost, "/uploads", handlers.CreateUpload(app))
	router.HandlerFunc(http.MethodPost, "/uploads/:id/complete", handlers.CompleteUpload(app))
	router.HandlerFunc(http.MethodGet, "/images", handlers.GetImages(app))
	router.HandlerFunc(http.MethodGet, "/image/:id", handlers.GetImageByID(app))
	router.HandlerFunc(http.MethodDelete, "/image/:id", handlers.DeleteImage(app))

	return mw.RecoverPanic(mw.EnableCORS(router))
}
