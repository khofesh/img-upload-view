package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/khofesh/img-upload-view/internal/config"
	"github.com/khofesh/img-upload-view/internal/data"
	"github.com/khofesh/img-upload-view/internal/reqres"
)

type envelope map[string]any

const MaxUploadSize = 10 << 20

func GetImages(app *config.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, err := reqres.ReadLimitParam(r)
		if err != nil {
			app.ErrorResponse.BadRequestResponse(w, r, err)
			return
		}

		offset, err := reqres.ReadOffsetParam(r)
		if err != nil {
			app.ErrorResponse.BadRequestResponse(w, r, err)
			return
		}

		if limit < 1 {
			limit = 20
		}
		if limit > 20 {
			limit = 20
		}

		images, totalCount, err := app.Models.Image.GetAll(limit, offset)
		if err != nil {
			app.ErrorResponse.ServerErrorResponse(w, r, fmt.Errorf("unable to retrieve images: %v", err))
			return
		}

		for _, image := range images {
			if err := fillImageURL(app, r, image); err != nil {
				app.ErrorResponse.ServerErrorResponse(w, r, err)
				return
			}
		}

		response := envelope{
			"images": images,
			"metadata": envelope{
				"total_count": totalCount,
				"limit":       limit,
				"offset":      offset,
				"has_more":    offset+limit < totalCount,
			},
		}

		err = reqres.WriteJSON(w, http.StatusOK, response, nil)
		if err != nil {
			app.ErrorResponse.ServerErrorResponse(w, r, err)
		}
	}
}

func GetImageByID(app *config.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		imageId, err := reqres.ReadIDParam(r)
		if err != nil {
			app.ErrorResponse.BadRequestResponse(w, r, err)
			return
		}

		image, err := app.Models.Image.GetByID(imageId)
		if err != nil {
			if err.Error() == "record not found" {
				app.ErrorResponse.NotFoundResponse(w, r)
				return
			}
			app.ErrorResponse.ServerErrorResponse(w, r, fmt.Errorf("unable to retrieve image: %v", err))
			return
		}

		if image.Status != data.StatusReady {
			app.ErrorResponse.NotFoundResponse(w, r)
			return
		}

		if err := fillImageURL(app, r, image); err != nil {
			app.ErrorResponse.ServerErrorResponse(w, r, err)
			return
		}

		response := envelope{
			"image": image,
		}

		err = reqres.WriteJSON(w, http.StatusOK, response, nil)
		if err != nil {
			app.ErrorResponse.ServerErrorResponse(w, r, err)
		}
	}
}

func DeleteImage(app *config.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		imageId, err := reqres.ReadIDParam(r)
		if err != nil {
			app.ErrorResponse.BadRequestResponse(w, r, err)
			return
		}

		// get image metadata first
		image, err := app.Models.Image.GetByID(imageId)
		if err != nil {
			if err.Error() == "record not found" {
				app.ErrorResponse.NotFoundResponse(w, r)
				return
			}
			app.ErrorResponse.ServerErrorResponse(w, r, fmt.Errorf("unable to retrieve image: %v", err))
			return
		}

		// delete from db
		err = app.Models.Image.Delete(imageId)
		if err != nil {
			if err.Error() == "record not found" {
				app.ErrorResponse.NotFoundResponse(w, r)
				return
			}
			app.ErrorResponse.ServerErrorResponse(w, r, fmt.Errorf("unable to delete image from database: %v", err))
			return
		}

		// delete the object; a missing object is not an error
		if err := app.Storage.Delete(r.Context(), image.ObjectKey); err != nil {
			app.Logger.Warn().Err(err).Str("object_key", image.ObjectKey).Msg("failed to delete object")
		}

		response := envelope{
			"message": "Image deleted successfully",
			"deleted_image": envelope{
				"id":                image.ID,
				"filename":          image.Filename,
				"original_filename": image.OriginalFilename,
			},
		}

		err = reqres.WriteJSON(w, http.StatusOK, response, nil)
		if err != nil {
			app.ErrorResponse.ServerErrorResponse(w, r, err)
		}
	}
}

func generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	timestamp := time.Now().Unix()

	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomString := hex.EncodeToString(randomBytes)

	return fmt.Sprintf("%d_%s%s", timestamp, randomString, ext)
}

func validateImageFile(size int64, contentType string) error {
	if size > MaxUploadSize {
		return fmt.Errorf("file size exceeds 10MB limit")
	}

	if contentType != "image/jpeg" && contentType != "image/jpg" {
		return fmt.Errorf("only JPEG images are allowed")
	}

	return nil
}
