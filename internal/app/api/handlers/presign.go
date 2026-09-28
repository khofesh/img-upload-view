package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/khofesh/img-upload-view/internal/config"
	"github.com/khofesh/img-upload-view/internal/data"
	"github.com/khofesh/img-upload-view/internal/reqres"
	"github.com/khofesh/img-upload-view/internal/storage"
)

const maxPresignRequestSize = 1 << 10 // 1 KB

type createUploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

// CreateUpload handles POST /uploads. It validates the declared file metadata,
// stores a pending row and returns a presigned PUT URL for direct-to-storage upload.
func CreateUpload(app *config.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxPresignRequestSize)

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		var req createUploadRequest
		if err := decoder.Decode(&req); err != nil {
			app.ErrorResponse.BadRequestResponse(w, r, fmt.Errorf("invalid request body: %v", err))
			return
		}

		if err := validateImageFile(req.Size, req.ContentType); err != nil {
			app.ErrorResponse.BadRequestResponse(w, r, err)
			return
		}

		if req.Size <= 0 {
			app.ErrorResponse.BadRequestResponse(w, r, errors.New("file size must be greater than 0"))
			return
		}

		// Clients never choose the key, so they cannot overwrite objects.
		key := "images/" + generateUniqueFilename(forceJPGExtension(req.Filename))

		image := &data.Image{
			Filename:         filepath.Base(key),
			OriginalFilename: req.Filename,
			ObjectKey:        key,
			FileSize:         req.Size,
			ContentType:      req.ContentType,
			UploadTimestamp:  time.Now(),
		}

		if err := app.Models.Image.InsertPending(image); err != nil {
			app.ErrorResponse.ServerErrorResponse(w, r, fmt.Errorf("unable to save image metadata: %v", err))
			return
		}

		presigned, err := app.Storage.PresignPut(r.Context(), key, req.ContentType, req.Size)
		if err != nil {
			if delErr := app.Models.Image.Delete(image.ID); delErr != nil {
				app.Logger.Warn().Err(delErr).Int64("image_id", image.ID).Msg("unable to roll back pending image")
			}
			app.ErrorResponse.ServerErrorResponse(w, r, fmt.Errorf("unable to create upload URL: %v", err))
			return
		}

		response := envelope{
			"id": image.ID,
			"upload": envelope{
				"url":        presigned.URL,
				"method":     presigned.Method,
				"headers":    presigned.Headers,
				"expires_at": presigned.ExpiresAt.UTC().Format(time.RFC3339),
			},
		}

		if err := reqres.WriteJSON(w, http.StatusCreated, response, nil); err != nil {
			app.ErrorResponse.ServerErrorResponse(w, r, err)
		}
	}
}

// CompleteUpload handles POST /uploads/:id/complete. It verifies the uploaded object
// with HeadObject and marks the row ready.
func CompleteUpload(app *config.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		imageID, err := reqres.ReadIDParam(r)
		if err != nil {
			app.ErrorResponse.BadRequestResponse(w, r, err)
			return
		}

		image, err := app.Models.Image.GetByID(imageID)
		if err != nil {
			if err.Error() == "record not found" {
				app.ErrorResponse.NotFoundResponse(w, r)
				return
			}
			app.ErrorResponse.ServerErrorResponse(w, r, fmt.Errorf("unable to retrieve image: %v", err))
			return
		}

		// Idempotent: a ready row can be completed again.
		if image.Status == data.StatusReady {
			if err := fillImageURL(app, r, image); err != nil {
				app.ErrorResponse.ServerErrorResponse(w, r, err)
				return
			}
			writeImageResponse(app, w, r, image)
			return
		}

		info, err := app.Storage.Head(r.Context(), image.ObjectKey)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				app.ErrorResponse.UnprocessableEntityResponse(w, r, errors.New("upload not found in storage"))
				return
			}
			app.ErrorResponse.ServerErrorResponse(w, r, fmt.Errorf("unable to verify uploaded object: %v", err))
			return
		}

		if err := validateImageFile(info.Size, info.ContentType); err != nil {
			deleteInvalidObject(app, r, image.ObjectKey)
			app.ErrorResponse.UnprocessableEntityResponse(w, r, err)
			return
		}

		if info.Size != image.FileSize {
			deleteInvalidObject(app, r, image.ObjectKey)
			app.ErrorResponse.UnprocessableEntityResponse(w, r, errors.New("uploaded object size does not match the declared size"))
			return
		}

		if err := app.Models.Image.MarkReady(image.ID, info.Size, info.ContentType); err != nil {
			if err.Error() == "record not found" {
				app.ErrorResponse.NotFoundResponse(w, r)
				return
			}
			app.ErrorResponse.ServerErrorResponse(w, r, fmt.Errorf("unable to mark image ready: %v", err))
			return
		}

		image.Status = data.StatusReady
		image.FileSize = info.Size
		image.ContentType = info.ContentType

		if err := fillImageURL(app, r, image); err != nil {
			app.ErrorResponse.ServerErrorResponse(w, r, err)
			return
		}

		writeImageResponse(app, w, r, image)
	}
}

func deleteInvalidObject(app *config.Application, r *http.Request, key string) {
	if err := app.Storage.Delete(r.Context(), key); err != nil {
		app.Logger.Warn().Err(err).Str("object_key", key).Msg("unable to delete invalid object")
	}
}

func fillImageURL(app *config.Application, r *http.Request, image *data.Image) error {
	url, err := app.Storage.PresignGet(r.Context(), image.ObjectKey)
	if err != nil {
		return fmt.Errorf("unable to create image URL: %v", err)
	}
	image.URL = url
	return nil
}

func writeImageResponse(app *config.Application, w http.ResponseWriter, r *http.Request, image *data.Image) {
	response := envelope{
		"message": "Image upload completed successfully",
		"image":   image,
	}

	if err := reqres.WriteJSON(w, http.StatusOK, response, nil); err != nil {
		app.ErrorResponse.ServerErrorResponse(w, r, err)
	}
}

func forceJPGExtension(filename string) string {
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	return base + ".jpg"
}
