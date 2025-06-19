package api

import (
	"net/http"
	"os"

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
	router.HandlerFunc(http.MethodGet, "/images", handlers.GetImages(app))
	router.HandlerFunc(http.MethodGet, "/image/:id", handlers.GetImageByID(app))

	// serving files for development
	if app.Config.Env == "local" {
		uploadDir := os.Getenv("UPLOAD_DIR")
		if uploadDir == "" {
			uploadDir = "./upload" // Default for local development
		}

		os.MkdirAll(uploadDir, 0755)
		router.ServeFiles("/images/*filepath", http.Dir(uploadDir))
	}

	return mw.RecoverPanic(mw.EnableCORS(router))
}
