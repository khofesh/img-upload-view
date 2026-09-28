//go:build integration

package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func newIntegrationStore(t *testing.T) *S3Store {
	t.Helper()

	endpoint := os.Getenv("STORAGE_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:9000"
	}

	logger := zerolog.Nop()
	store := NewS3Store(Options{
		Endpoint:       endpoint,
		PublicEndpoint: endpoint,
		Region:         "us-east-1",
		Bucket:         "images-integration-test",
		AccessKey:      "minioadmin",
		SecretKey:      "minioadmin",
		UsePathStyle:   true,
		PresignPutTTL:  5 * time.Minute,
		PresignGetTTL:  5 * time.Minute,
	}, &logger)

	if err := store.EnsureBucket(context.Background()); err != nil {
		t.Fatalf("EnsureBucket: %v", err)
	}
	return store
}

func TestS3StoreLifecycle(t *testing.T) {
	store := newIntegrationStore(t)
	ctx := context.Background()
	key := "images/integration.jpg"
	body := bytes.Repeat([]byte("A"), 1000)

	presigned, err := store.PresignPut(ctx, key, "image/jpeg", int64(len(body)))
	if err != nil {
		t.Fatalf("PresignPut: %v", err)
	}

	put(t, presigned.URL, "image/jpeg", body, http.StatusOK)

	info, err := store.Head(ctx, key)
	if err != nil {
		t.Fatalf("Head: %v", err)
	}
	if info.Size != int64(len(body)) {
		t.Errorf("size = %d, want %d", info.Size, len(body))
	}

	getURL, err := store.PresignGet(ctx, key)
	if err != nil {
		t.Fatalf("PresignGet: %v", err)
	}
	resp, err := http.Get(getURL)
	if err != nil {
		t.Fatalf("GET presigned: %v", err)
	}
	defer resp.Body.Close()
	got, _ := io.ReadAll(resp.Body)
	if !bytes.Equal(got, body) {
		t.Errorf("GET body mismatch")
	}

	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Head(ctx, key); !errors.Is(err, ErrNotFound) {
		t.Errorf("Head after delete = %v, want ErrNotFound", err)
	}
}

func TestS3StoreRejectsOversizedBody(t *testing.T) {
	store := newIntegrationStore(t)
	ctx := context.Background()
	key := "images/oversize.jpg"

	presigned, err := store.PresignPut(ctx, key, "image/jpeg", 1000)
	if err != nil {
		t.Fatalf("PresignPut: %v", err)
	}

	put(t, presigned.URL, "image/jpeg", bytes.Repeat([]byte("A"), 2000), http.StatusForbidden)
}

func put(t *testing.T, url, contentType string, body []byte, wantStatus int) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = int64(len(body))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != wantStatus {
		msg, _ := io.ReadAll(resp.Body)
		t.Fatalf("PUT status = %d, want %d (body: %s)", resp.StatusCode, wantStatus, msg)
	}
}
