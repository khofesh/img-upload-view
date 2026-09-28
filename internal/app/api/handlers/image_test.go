package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/khofesh/img-upload-view/internal/data"
	"github.com/khofesh/img-upload-view/internal/storage"
)

func TestValidateImageFile(t *testing.T) {
	tests := []struct {
		name        string
		size        int64
		contentType string
		wantErr     bool
	}{
		{name: "jpeg", size: 1000, contentType: "image/jpeg"},
		{name: "jpg", size: 1000, contentType: "image/jpg"},
		{name: "at limit", size: MaxUploadSize, contentType: "image/jpeg"},
		{name: "over limit", size: MaxUploadSize + 1, contentType: "image/jpeg", wantErr: true},
		{name: "png", size: 1000, contentType: "image/png", wantErr: true},
		{name: "empty type", size: 1000, contentType: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateImageFile(tt.size, tt.contentType)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateImageFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteImage(t *testing.T) {
	model := newFakeImageModel()
	store := newFakeObjectStore()
	key := "images/1_abcdef.jpg"
	model.images[1] = &data.Image{
		ID:              1,
		ObjectKey:       key,
		Status:          data.StatusReady,
		UploadTimestamp: time.Now(),
	}
	store.objects[key] = storage.ObjectInfo{Size: 1000, ContentType: "image/jpeg"}

	app := newTestApp(model, store)

	rec := httptest.NewRecorder()
	params := httprouter.Params{{Key: "id", Value: "1"}}
	req := requestWithParams(http.MethodDelete, "/image/1", nil, params)
	DeleteImage(app)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}
	if _, err := model.GetByID(1); err == nil {
		t.Errorf("expected row to be deleted")
	}
	if len(store.deleted) != 1 || store.deleted[0] != key {
		t.Errorf("deleted objects = %v, want [%s]", store.deleted, key)
	}
}

func TestGetImageByIDHidesPending(t *testing.T) {
	model := newFakeImageModel()
	store := newFakeObjectStore()
	model.images[1] = &data.Image{ID: 1, ObjectKey: "images/x.jpg", Status: data.StatusPending}
	app := newTestApp(model, store)

	rec := httptest.NewRecorder()
	params := httprouter.Params{{Key: "id", Value: "1"}}
	req := requestWithParams(http.MethodGet, "/image/1", nil, params)
	GetImageByID(app)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
