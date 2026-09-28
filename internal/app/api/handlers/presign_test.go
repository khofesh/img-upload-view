package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/khofesh/img-upload-view/internal/config"
	"github.com/khofesh/img-upload-view/internal/data"
	"github.com/khofesh/img-upload-view/internal/storage"
	errres "github.com/khofesh/img-upload-view/pkg/errors"
	"github.com/rs/zerolog"
)

func newTestApp(model *fakeImageModel, store *fakeObjectStore) *config.Application {
	logger := zerolog.Nop()
	return &config.Application{
		Logger:        &logger,
		Config:        &config.Config{},
		Models:        data.Models{Image: model},
		Storage:       store,
		ErrorResponse: errres.NewErrorResponse(&logger),
	}
}

func requestWithParams(method, target string, body *bytes.Reader, params httprouter.Params) *http.Request {
	var req *http.Request
	if body == nil {
		req = httptest.NewRequest(method, target, nil)
	} else {
		req = httptest.NewRequest(method, target, body)
	}
	ctx := context.WithValue(req.Context(), httprouter.ParamsKey, params)
	return req.WithContext(ctx)
}

func TestCreateUpload(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		presignErr error
		wantStatus int
	}{
		{
			name:       "happy path",
			body:       `{"filename":"cat.jpg","content_type":"image/jpeg","size":1000}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid content type",
			body:       `{"filename":"cat.png","content_type":"image/png","size":1000}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "zero size",
			body:       `{"filename":"cat.jpg","content_type":"image/jpeg","size":0}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "oversize",
			body:       `{"filename":"cat.jpg","content_type":"image/jpeg","size":10485761}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown field",
			body:       `{"filename":"cat.jpg","content_type":"image/jpeg","size":1000,"key":"evil"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "presign failure",
			body:       `{"filename":"cat.jpg","content_type":"image/jpeg","size":1000}`,
			presignErr: errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newFakeImageModel()
			store := newFakeObjectStore()
			store.presignPutErr = tt.presignErr
			app := newTestApp(model, store)

			rec := httptest.NewRecorder()
			req := requestWithParams(http.MethodPost, "/uploads", bytes.NewReader([]byte(tt.body)), nil)
			CreateUpload(app)(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantStatus != http.StatusCreated {
				return
			}

			var resp struct {
				ID     int64 `json:"id"`
				Upload struct {
					URL     string            `json:"url"`
					Method  string            `json:"method"`
					Headers map[string]string `json:"headers"`
				} `json:"upload"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("invalid json response: %v", err)
			}
			if resp.ID != 1 {
				t.Errorf("id = %d, want 1", resp.ID)
			}
			if resp.Upload.Method != "PUT" {
				t.Errorf("method = %q, want PUT", resp.Upload.Method)
			}
			if !strings.Contains(resp.Upload.URL, "images/") {
				t.Errorf("url = %q, want it to contain images/", resp.Upload.URL)
			}
			if resp.Upload.Headers["Content-Type"] != "image/jpeg" {
				t.Errorf("content-type header = %q", resp.Upload.Headers["Content-Type"])
			}

			image, err := model.GetByID(1)
			if err != nil {
				t.Fatalf("pending row missing: %v", err)
			}
			if image.Status != data.StatusPending {
				t.Errorf("status = %q, want pending", image.Status)
			}
			if !strings.HasPrefix(image.ObjectKey, "images/") || !strings.HasSuffix(image.ObjectKey, ".jpg") {
				t.Errorf("object key = %q", image.ObjectKey)
			}
		})
	}

	t.Run("presign failure rolls back pending row", func(t *testing.T) {
		model := newFakeImageModel()
		store := newFakeObjectStore()
		store.presignPutErr = errors.New("boom")
		app := newTestApp(model, store)

		rec := httptest.NewRecorder()
		body := bytes.NewReader([]byte(`{"filename":"cat.jpg","content_type":"image/jpeg","size":1000}`))
		CreateUpload(app)(rec, requestWithParams(http.MethodPost, "/uploads", body, nil))

		if len(model.images) != 0 {
			t.Errorf("expected pending row to be rolled back, got %d rows", len(model.images))
		}
	})
}

func TestCompleteUpload(t *testing.T) {
	const key = "images/123_abcdef.jpg"

	seedPending := func(model *fakeImageModel) {
		model.images[1] = &data.Image{
			ID:              1,
			ObjectKey:       key,
			FileSize:        1000,
			ContentType:     "image/jpeg",
			Status:          data.StatusPending,
			UploadTimestamp: time.Now(),
		}
	}

	tests := []struct {
		name        string
		seed        func(*fakeImageModel)
		object      *storage.ObjectInfo
		wantStatus  int
		wantDeleted bool
		wantReady   bool
	}{
		{
			name:       "happy path",
			seed:       seedPending,
			object:     &storage.ObjectInfo{Size: 1000, ContentType: "image/jpeg"},
			wantStatus: http.StatusOK,
			wantReady:  true,
		},
		{
			name:       "missing object",
			seed:       seedPending,
			object:     nil,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:        "oversize object",
			seed:        seedPending,
			object:      &storage.ObjectInfo{Size: MaxUploadSize + 1, ContentType: "image/jpeg"},
			wantStatus:  http.StatusUnprocessableEntity,
			wantDeleted: true,
		},
		{
			name:        "wrong type",
			seed:        seedPending,
			object:      &storage.ObjectInfo{Size: 1000, ContentType: "text/plain"},
			wantStatus:  http.StatusUnprocessableEntity,
			wantDeleted: true,
		},
		{
			name:        "size mismatch",
			seed:        seedPending,
			object:      &storage.ObjectInfo{Size: 2000, ContentType: "image/jpeg"},
			wantStatus:  http.StatusUnprocessableEntity,
			wantDeleted: true,
		},
		{
			name:       "unknown id",
			seed:       func(*fakeImageModel) {},
			object:     &storage.ObjectInfo{Size: 1000, ContentType: "image/jpeg"},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newFakeImageModel()
			store := newFakeObjectStore()
			tt.seed(model)
			if tt.object != nil {
				store.objects[key] = *tt.object
			}
			app := newTestApp(model, store)

			rec := httptest.NewRecorder()
			params := httprouter.Params{{Key: "id", Value: "1"}}
			req := requestWithParams(http.MethodPost, "/uploads/1/complete", nil, params)
			CompleteUpload(app)(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantDeleted && len(store.deleted) == 0 {
				t.Errorf("expected invalid object to be deleted")
			}

			if tt.wantReady {
				image, _ := model.GetByID(1)
				if image.Status != data.StatusReady {
					t.Errorf("status = %q, want ready", image.Status)
				}
				if image.URL == "" {
					t.Errorf("expected image URL to be filled")
				}
			}
		})
	}

	t.Run("complete is idempotent", func(t *testing.T) {
		model := newFakeImageModel()
		store := newFakeObjectStore()
		seedPending(model)
		store.objects[key] = storage.ObjectInfo{Size: 1000, ContentType: "image/jpeg"}
		app := newTestApp(model, store)

		params := httprouter.Params{{Key: "id", Value: "1"}}
		for i := 0; i < 2; i++ {
			rec := httptest.NewRecorder()
			req := requestWithParams(http.MethodPost, "/uploads/1/complete", nil, params)
			CompleteUpload(app)(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("call %d: status = %d, want 200 (body: %s)", i+1, rec.Code, rec.Body.String())
			}
		}
	})
}
