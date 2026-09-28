# Plan — Object Storage with Presigned Uploads

- PRD: [object-storage-presigned-upload-prd.md](./object-storage-presigned-upload-prd.md)
- Backend: MinIO (dev + compose prod), S3 API via `github.com/aws/aws-sdk-go-v2`
- Approach: add the new flow alongside `POST /upload`, switch the frontend, then delete the old path.

## Phase 0 — Spike (half a day)

Answer PRD Q1, Q5 and Q6 before writing production code.

- [ ] Pick and pin a MinIO image tag that is available (check the 2025 distribution change). If
      none is usable, switch this plan to LocalStack. Only the compose/config steps change.
- [ ] Throwaway `main.go`: `PresignPutObject` with `ContentType` + `ContentLength` against
      `http://localhost:9000`, `UsePathStyle: true`. Check with curl:
  - [ ] matching body → 200
  - [ ] larger body → rejected (`SignatureDoesNotMatch`)
  - [ ] wrong `Content-Type` → rejected
- [ ] Browser `fetch(url, {method: "PUT"})` from `localhost:5173` works (CORS preflight passes).

## Phase 1 — Infrastructure & config

**`compose.dev.yaml`** — add:

```yaml
minio:
  image: minio/minio:<pinned-tag>
  command: server /data --console-address ":9001"
  environment:
    - MINIO_ROOT_USER=minioadmin
    - MINIO_ROOT_PASSWORD=minioadmin
    - MINIO_API_CORS_ALLOW_ORIGIN=http://localhost:5173,http://localhost:3000
  ports: ["9000:9000", "9001:9001"]
  volumes: [minio_data:/data]
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
```

**`compose.yaml`** — same service on `api-network`:

- CORS origin `http://localhost`.
- `api-service` `depends_on: minio (service_healthy)`.
- Remove the `uploaded_images` volume, the `UPLOAD_DIR` env var, and the nginx image mount.

**`internal/config/config.go`** — add:

```go
Storage struct {
    Endpoint       string        `yaml:"endpoint"`       // API → storage, e.g. http://minio:9000
    PublicEndpoint string        `yaml:"publicEndpoint"` // browser → storage, used for presigning
    Region         string        `yaml:"region"`
    Bucket         string        `yaml:"bucket"`
    AccessKey      string        `yaml:"accessKey"`
    SecretKey      string        `yaml:"secretKey"`
    UsePathStyle   bool          `yaml:"usePathStyle"`
    CreateBucket   bool          `yaml:"createBucket"`
    PresignPutTTL  time.Duration `yaml:"presignPutTTL"`  // "5m"
    PresignGetTTL  time.Duration `yaml:"presignGetTTL"`  // "15m"
} `yaml:"storage"`
```

- [ ] Add the `storage:` block to `config.dev.yaml` (endpoint = publicEndpoint = `http://localhost:9000`)
      and to `configs/config.yaml` (endpoint `http://minio:9000`, publicEndpoint `http://localhost:9000`).
- [ ] Set defaults for zero TTLs (5m / 15m) in one place after config load.

## Phase 2 — Storage adapter

New package `internal/storage`:

```go
type PresignedRequest struct {
    URL       string
    Method    string
    Headers   map[string]string
    ExpiresAt time.Time
}

type ObjectInfo struct {
    Size        int64
    ContentType string
}

var ErrNotFound = errors.New("object not found")

type ObjectStore interface {
    PresignPut(ctx context.Context, key, contentType string, size int64) (PresignedRequest, error)
    PresignGet(ctx context.Context, key string) (string, error)
    Head(ctx context.Context, key string) (ObjectInfo, error) // ErrNotFound when missing
    Delete(ctx context.Context, key string) error              // nil when already gone
    EnsureBucket(ctx context.Context) error
}
```

- [ ] `s3.go`: `S3Store` builds two `*s3.Client`s with static credentials, `BaseEndpoint`,
      `UsePathStyle`:
  - `internal` for Head, Delete and EnsureBucket
  - `public` wrapped in `s3.NewPresignClient`
- [ ] Map `types.NotFound` / `NoSuchKey` to `ErrNotFound` in `Head`.
- [ ] `EnsureBucket`: `HeadBucket`, then `CreateBucket` on 404. Try `PutBucketCors` and ignore
      `NotImplemented` (MinIO).
- [ ] `go get github.com/aws/aws-sdk-go-v2/{config,credentials,service/s3}`, then `make tidy`.
- [ ] Wire it up: add `Storage storage.ObjectStore` to `config.Application`. In `cmd/api/main.go`,
      construct it after the DB and call `EnsureBucket` when `cfg.Storage.CreateBucket`.

## Phase 3 — Schema & data layer

- [ ] `configs/postgres-init/01-init.sql`: add `object_key VARCHAR(500) NOT NULL UNIQUE` and
      `status VARCHAR(16) NOT NULL DEFAULT 'pending'`, and drop `url`. Replace the timestamp index with
      `CREATE INDEX ... ON images(upload_timestamp DESC) WHERE status = 'ready'`.
- [ ] Existing dev DBs (the init script won't rerun) have two options:
  - recreate the volume with `docker compose -f compose.dev.yaml down -v`
  - apply `docs/todo/sql/002-object-storage.sql` (an `ALTER TABLE` script shipped with this change)
    with `psql`
- [ ] `internal/data/image_data.go`:
  - `Image`: add `ObjectKey string \`json:"-"\``, `Status string \`json:"status"\``. Keep
`URL string \`json:"url"\`` but stop persisting it, and fill it in the handler.
  - `IImageModel`: add `InsertPending(*Image) error` and `MarkReady(id, size int64, contentType string) error`,
    plus `DeleteStalePending(olderThan time.Duration) ([]*Image, error)` for cleanup.
  - `GetAll`: add `WHERE status = 'ready'` to the count query and the list query. `GetByID`
    returns rows in any status; the handlers decide what to do with them.
  - Keep `"record not found"` as the not-found error text.

## Phase 4 — API handlers

New file `internal/app/api/handlers/presign.go`:

- [ ] `CreateUpload(app)` — `POST /uploads`
  - Decode JSON (limit body to 1 KB with `http.MaxBytesReader`, reject unknown fields).
  - Reuse `validateImageFile(size, contentType)`, and also reject `size <= 0`.
  - `key := "images/" + generateUniqueFilename(filename)`, forcing the `.jpg` extension.
  - `InsertPending`, then `Storage.PresignPut`. If presigning fails, delete the pending row.
  - Return 201 `{id, upload:{url, method, headers, expires_at}}`.
- [ ] `CompleteUpload(app)` — `POST /uploads/:id/complete`
  - `GetByID`: 404 if the row is missing. If it's already `ready`, return 200 with the image.
  - `Storage.Head(key)`:
    - `ErrNotFound` → 422 `"upload not found in storage"`.
    - Size mismatch, over the limit, or wrong type → `Storage.Delete`, then 422.
  - `MarkReady`, fill `image.URL` from `PresignGet`, return 200.
- [ ] `GetImages` / `GetImageByID`: fill `URL` for each image with `Storage.PresignGet`.
      `GetImageByID` returns 404 for `pending` rows.
- [ ] `DeleteImage`: replace `os.Remove` with `Storage.Delete(r.Context(), image.ObjectKey)`.
      Keep "delete the row, then the object; log a warning if the object delete fails".
- [ ] `routes.go`: register `POST /uploads` and `POST /uploads/:id/complete`.
- [ ] Nginx needs no change for the upload itself: the browser PUTs straight to `localhost:9000`.

## Phase 5 — Frontend

`web/app/routes/upload.tsx`, inside `handleUpload`:

- [ ] Step 1: `POST ${apiUrl}/uploads` with JSON `{filename, content_type: file.type, size: file.size}`.
- [ ] Step 2: `PUT upload.url`, `body: file`, `headers: upload.headers`. Use `XMLHttpRequest` if an
      upload progress bar is wanted, otherwise `fetch`.
- [ ] Step 3: `POST ${apiUrl}/uploads/${id}/complete`.
- [ ] Show a clear error for each step. S3 errors come back as XML, so show a generic message for
      step 2 failures.
- [ ] `gallery.tsx`: no change expected. `image.url` is now absolute. Check that images still load
      after the presigned GET TTL expires (refetch on error or on focus, if needed).
- [ ] `npm run typecheck`.

## Phase 6 — Remove the local-disk path

- [ ] Delete `UploadImage` and the `POST /upload` route.
- [ ] Delete the `UploadDir` constant and all reads of `UPLOAD_DIR`.
- [ ] Delete the `env == "local"` static file server in `routes.go`.
- [ ] `configs/nginx.conf`: remove `location /images/` and `client_max_body_size 10M`.
- [ ] `.gitignore`: `upload/` can stay, since it's harmless.

## Phase 7 — Cleanup job

- [ ] `cmd/cli` / `internal/app/cli`: replace the stub with a `cleanup` subcommand. It reads the same
      config, calls `DeleteStalePending(1h)`, and runs `Storage.Delete` on each returned key.
- [ ] Add a `make run/cli/cleanup` target.

## Phase 8 — Tests

The project has no tests yet. Add:

- [ ] `handlers/presign_test.go`: table-driven tests with a fake `ObjectStore` and a fake
      `IImageModel`:
  - invalid type, size 0, oversize
  - happy path
  - complete when the object is missing, oversize, or of the wrong type
  - complete called twice
- [ ] `validateImageFile` table test.
- [ ] `internal/storage/s3_integration_test.go` (`//go:build integration`): run against the compose
      MinIO. Presign PUT, upload with `net/http`, Head, oversized body rejected, Delete.
- [ ] `make audit` is green.

## Phase 9 — Docs

- [ ] `README.md`:
  - new dev steps (MinIO console at `:9001`)
  - curl examples for the three-step flow
  - fix the `compose-dev.yaml` typo and the `/images/1` → `/image/1` example
- [ ] `AGENTS.md`: update the project summary, key facts (storage config, presign flow, no
      `UPLOAD_DIR`), the routes table and the commands.
- [ ] Rebuild the graph: `/graphify . --update`.
- [ ] Move this PRD and plan to `docs/done/`.

## Verification checklist

```bash
docker compose -f compose.dev.yaml up -d
make run/api
cd web && VITE_API_URL=http://localhost:8080 npm run dev
```

- [ ] Upload a JPEG in the UI. The object shows up in the MinIO console, the row is `ready`, and the
      gallery shows it.
- [ ] Upload the file from `dummy-jpeg/` (> 10 MB): the frontend blocks it.
- [ ] Oversize bypass with curl: presign with `size=1000`, PUT 11 MB → storage rejects it.
- [ ] Presign, then skip the PUT and call complete → 422. After `cli cleanup`, the row is gone.
- [ ] Delete from the gallery: the row and the object are both gone.
- [ ] `docker compose -f compose.yaml up --build`: the same flow works through `http://localhost`.

## Order & sizing

| Phase             | Depends on | Size |
| ----------------- | ---------- | ---- |
| 0 Spike           | —          | S    |
| 1 Infra/config    | 0          | S    |
| 2 Storage adapter | 1          | M    |
| 3 Schema/data     | —          | S    |
| 4 Handlers        | 2, 3       | M    |
| 5 Frontend        | 4          | S    |
| 6 Remove old path | 5          | S    |
| 7 Cleanup CLI     | 3, 2       | S    |
| 8 Tests           | 2, 4       | M    |
| 9 Docs            | all        | S    |

Phases 2 and 3 can run in parallel. Phases 1–5 can ship as one PR, with phases 6–9 as a follow-up
PR, so the old endpoint remains as a fallback until the new flow is verified.
