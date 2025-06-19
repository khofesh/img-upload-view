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
		err = validateImageFile(file, fileHeader.Size)
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
			"image": envelope{
				"id":                imageData.ID,
				"filename":          imageData.Filename,
				"original_filename": imageData.OriginalFilename,
				"url":               imageData.URL,
				"file_size":         imageData.FileSize,
				"content_type":      imageData.ContentType,
				"upload_timestamp":  imageData.UploadTimestamp,
			},
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

		images, totalCount, err := app.Models.Image.GetAll(limit, offset)
		if err != nil {
			app.ErrorResponse.ServerErrorResponse(w, r, fmt.Errorf("unable to retrieve images: %v", err))
			return
		}

		imageList := make([]envelope, len(images))
		for i, img := range images {
			imageList[i] = envelope{
				"id":                img.ID,
				"filename":          img.Filename,
				"original_filename": img.OriginalFilename,
				"url":               img.URL,
				"file_size":         img.FileSize,
				"content_type":      img.ContentType,
				"upload_timestamp":  img.UploadTimestamp,
			}
		}

		response := envelope{
			"images": imageList,
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

		// TODO: use json.marshal
		response := envelope{
			"image": envelope{
				"id":                image.ID,
				"filename":          image.Filename,
				"original_filename": image.OriginalFilename,
				"url":               image.URL,
				"file_size":         image.FileSize,
				"content_type":      image.ContentType,
				"upload_timestamp":  image.UploadTimestamp,
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

func validateImageFile(file io.Reader, size int64) error {
	if size > MaxUploadSize {
		return fmt.Errorf("file size exceeds 10MB limit")
	}

	return nil
}
