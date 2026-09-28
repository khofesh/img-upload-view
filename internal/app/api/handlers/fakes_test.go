package handlers

import (
	"context"
	"errors"
	"time"

	"github.com/khofesh/img-upload-view/internal/data"
	"github.com/khofesh/img-upload-view/internal/storage"
)

type fakeImageModel struct {
	images    map[int64]*data.Image
	nextID    int64
	insertErr error
	markErr   error
}

func newFakeImageModel() *fakeImageModel {
	return &fakeImageModel{images: map[int64]*data.Image{}, nextID: 1}
}

func (m *fakeImageModel) InsertPending(image *data.Image) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	image.ID = m.nextID
	m.nextID++
	image.Status = data.StatusPending
	stored := *image
	m.images[image.ID] = &stored
	return nil
}

func (m *fakeImageModel) MarkReady(id, size int64, contentType string) error {
	if m.markErr != nil {
		return m.markErr
	}
	image, ok := m.images[id]
	if !ok {
		return errors.New("record not found")
	}
	image.Status = data.StatusReady
	image.FileSize = size
	image.ContentType = contentType
	return nil
}

func (m *fakeImageModel) GetAll(limit, offset int64) ([]*data.Image, int64, error) {
	images := []*data.Image{}
	var total int64
	for _, image := range m.images {
		if image.Status != data.StatusReady {
			continue
		}
		total++
		if int64(len(images)) >= limit {
			continue
		}
		images = append(images, image)
	}
	return images, total, nil
}

func (m *fakeImageModel) GetByID(id int64) (*data.Image, error) {
	image, ok := m.images[id]
	if !ok {
		return nil, errors.New("record not found")
	}
	return image, nil
}

func (m *fakeImageModel) Delete(id int64) error {
	if _, ok := m.images[id]; !ok {
		return errors.New("record not found")
	}
	delete(m.images, id)
	return nil
}

func (m *fakeImageModel) DeleteStalePending(olderThan time.Duration) ([]*data.Image, error) {
	cutoff := time.Now().Add(-olderThan)
	var deleted []*data.Image
	for id, image := range m.images {
		if image.Status == data.StatusPending && image.UploadTimestamp.Before(cutoff) {
			deleted = append(deleted, image)
			delete(m.images, id)
		}
	}
	return deleted, nil
}

type fakeObjectStore struct {
	objects       map[string]storage.ObjectInfo
	presignPutErr error
	presignGetErr error
	headErr       error
	deleted       []string
}

func newFakeObjectStore() *fakeObjectStore {
	return &fakeObjectStore{objects: map[string]storage.ObjectInfo{}}
}

func (s *fakeObjectStore) PresignPut(_ context.Context, key, contentType string, size int64) (storage.PresignedRequest, error) {
	if s.presignPutErr != nil {
		return storage.PresignedRequest{}, s.presignPutErr
	}
	return storage.PresignedRequest{
		URL:       "http://storage.local/" + key,
		Method:    "PUT",
		Headers:   map[string]string{"Content-Type": contentType},
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}, nil
}

func (s *fakeObjectStore) PresignGet(_ context.Context, key string) (string, error) {
	if s.presignGetErr != nil {
		return "", s.presignGetErr
	}
	return "http://storage.local/" + key + "?signed", nil
}

func (s *fakeObjectStore) Head(_ context.Context, key string) (storage.ObjectInfo, error) {
	if s.headErr != nil {
		return storage.ObjectInfo{}, s.headErr
	}
	info, ok := s.objects[key]
	if !ok {
		return storage.ObjectInfo{}, storage.ErrNotFound
	}
	return info, nil
}

func (s *fakeObjectStore) Delete(_ context.Context, key string) error {
	s.deleted = append(s.deleted, key)
	delete(s.objects, key)
	return nil
}

func (s *fakeObjectStore) EnsureBucket(_ context.Context) error { return nil }
