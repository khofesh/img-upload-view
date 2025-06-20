package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/khofesh/img-upload-view/internal/config"
	"github.com/khofesh/img-upload-view/internal/data"
	"github.com/khofesh/img-upload-view/internal/reqres"
)

type envelope map[string]any

const (
	MaxUploadSize = 10 << 20
	UploadDir     = "/app/uploads"
)

func UploadImage(app *config.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(MaxUploadSize)
		if err != nil {
			app.ErrorResponse.BadRequestResponse(w, r, fmt.Errorf("unable to parse form: %v", err))
			return
		}

		// get the file
		file, fileHeader, err := r.FormFile("image")
		if err != nil {
			app.ErrorResponse.BadRequestResponse(w, r, fmt.Errorf("unable to get image file: %v", err))
			return
		}
		defer file.Close()

		// validate
		err = validateImageFile(fileHeader.Size, fileHeader.Header.Get("Content-Type"))
		if err != nil {
			app.ErrorResponse.BadRequestResponse(w, r, err)
			return
		}

		// gen unique filename
		uniqueFilename := generateUniqueFilename(fileHeader.Filename)

		uploadDir := os.Getenv("UPLOAD_DIR")
		if uploadDir == "" {
			uploadDir = UploadDir
		}

		// NOTE: we could instead send the image data to GCP cloud storage
		// or AWS S3 bucket
		err = os.MkdirAll(uploadDir, 0755)
		if err != nil {
			app.ErrorResponse.ServerErrorResponse(w, r, fmt.Errorf("unable to create upload directory: %v", err))
			return
		}

		// destination file
		filePath := filepath.Join(uploadDir, uniqueFilename)
		destFile, err := os.Create(filePath)
		if err != nil {
			app.ErrorResponse.ServerErrorResponse(w, r, fmt.Errorf("unable to create destination file: %v", err))
			return
		}
		defer destFile.Close()

		// reset file pointer
		file.Seek(0, 0)

		// copy
		_, err = io.Copy(destFile, file)
		if err != nil {
			app.ErrorResponse.ServerErrorResponse(w, r, fmt.Errorf("unable to save file: %v", err))
			return
		}

		// construct image metadata
		imageData := &data.Image{
			Filename:         uniqueFilename,
			OriginalFilename: fileHeader.Filename,
			URL:              fmt.Sprintf("/images/%s", uniqueFilename),
			FileSize:         fileHeader.Size,
			ContentType:      fileHeader.Header.Get("Content-Type"),
			UploadTimestamp:  time.Now(),
		}

		err = app.Models.Image.Insert(imageData)
		if err != nil {
			// delete file if error during insert
			os.Remove(filePath)
			app.ErrorResponse.ServerErrorResponse(w, r, fmt.Errorf("unable to save image metadata: %v", err))
			return
		}

		response := envelope{
			"message": "Image uploaded successfully",
			"image":   imageData,
		}

		err = reqres.WriteJSON(w, http.StatusCreated, response, nil)

		if err != nil {
			app.ErrorResponse.ServerErrorResponse(w, r, err)
		}
	}
}

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

		if limit > 20 {
			limit = 20
		}

		images, totalCount, err := app.Models.Image.GetAll(limit, offset)
		if err != nil {
			app.ErrorResponse.ServerErrorResponse(w, r, fmt.Errorf("unable to retrieve images: %v", err))
			return
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

		// delete the physical file
		uploadDir := os.Getenv("UPLOAD_DIR")
		if uploadDir == "" {
			uploadDir = UploadDir
		}

		filePath := filepath.Join(uploadDir, image.Filename)
		err = os.Remove(filePath)
		if err != nil && !os.IsNotExist(err) {
			// maybe do some cleanup later for orphan file
			app.Logger.Warn().Msgf("failed to delete physical file %s: %v", filePath, err)
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
