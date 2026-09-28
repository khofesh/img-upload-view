# AGENTS.md

## Project

Image upload & gallery service. Go 1.24 JSON API + React Router 7 (SSR) frontend, PostgreSQL
for metadata, JPEGs stored in S3-compatible object storage (MinIO AIStor). The browser uploads
directly to storage with presigned URLs; image bytes never pass through the API.

```
cmd/api/                  API entry point (flag -config-path, default /etc/secrets/config.yaml)
cmd/cli/                  CLI entry point (cleanup subcommand)
internal/app/api/         server.go (http.Server + graceful shutdown), routes.go (httprouter)
internal/app/api/handlers/ image.go — list / get / delete; presign.go — create / complete upload
internal/app/bootstrap/   NewStorage — builds the ObjectStore from Config
internal/config/          Config (YAML) + Application (DI container passed to handlers)
internal/data/            Models, IImageModel interface, ImageModel (database/sql + lib/pq)
internal/db/              OpenDB — pool setup + ping
internal/middleware/      Generic Middlewares[T] with functional options: CORS, RecoverPanic
internal/reqres/          Request param readers, WriteJSON, WriteFile
internal/storage/         ObjectStore interface + S3Store (aws-sdk-go-v2), two S3 clients
pkg/errors/               ErrorResponse — JSON error envelope helpers
pkg/read-config/          YAML config loader
configs/                  prod config.yaml, nginx.conf, postgres-init/01-init.sql (schema)
docs/todo/sql/            ALTER TABLE migration scripts for existing dev databases
infra/terraform/minio/    Terraform: images bucket, img-api IAM user/policy, generated API config
web/                      React Router 7 + Vite + Tailwind 4 frontend (routes: home, upload, gallery)
dummy-jpeg/               Python script generating oversized test JPEGs
```

Key facts:

- No migrations tool. Schema lives in `configs/postgres-init/01-init.sql`, applied only on first
  Postgres container init. Schema changes require editing that file and recreating the volume.
- Handlers are closures: `func X(app *config.Application) http.HandlerFunc`. Dependencies go
  through `config.Application` (`Logger`, `Config`, `Models`, `Storage`, `ErrorResponse`).
- Data access goes through `data.Models.Image` (`IImageModel`). Add new queries to the interface
  and `ImageModel`.
- "Not found" is signalled by `errors.New("record not found")` and matched by string in
  handlers — keep that message stable.
- Uploads are a three-step presigned flow. `POST /uploads` validates the declared
  `{filename,content_type,size}` (JPEG only, max 10 MB `MaxUploadSize`), inserts a `pending` row
  and returns a presigned PUT. The browser PUTs bytes straight to storage, then
  `POST /uploads/:id/complete` runs `HeadObject`, verifies size/type and marks the row `ready`.
- Object keys are server-generated: `images/<unix>_<16 hex>.jpg`. The bucket is private;
  `image.url` is a presigned GET generated per request.
- `internal/storage.ObjectStore` abstracts storage (`PresignPut`, `PresignGet`, `Head`, `Delete`,
  `EnsureBucket`). `S3Store` keeps two S3 clients: an internal endpoint for Head/Delete and a
  `publicEndpoint` for presigning, because SigV4 signs the host. `Head` returns
  `storage.ErrNotFound` when missing.
- Config `storage:` block holds endpoint/publicEndpoint/bucket/credentials/TTLs. `presignPutTTL`
  and `presignGetTTL` default to 5m/15m via `Config.ApplyDefaults()`.
- `env: local` vs prod no longer changes file serving. In prod nginx proxies `/api/*` → API
  (prefix stripped) and `/` → frontend; the browser fetches images from `publicEndpoint`.
- Config is YAML only (`config.dev.yaml` locally, `configs/config.yaml` in prod). The `PORT` /
  `DB_DSN` env vars in `compose.yaml` are not read by the code.
- Terraform (`infra/terraform/minio/`) can provision the bucket and a least-privilege `img-api`
  user, writing a gitignored `config.tf.yaml` with `createBucket: false`. It is opt-in; the default
  `make run/api` path keeps using root credentials and `createBucket: true`.
- MinIO AIStor needs a license file mounted at `/minio.license`; compose reads
  `${MINIO_LICENSE_PATH:-./minio.license}`.
- Frontend calls `${VITE_API_URL || "/api"}`. Dev: `VITE_API_URL=http://localhost:8080`.
- Logging: zerolog (global `log.Logger`, JSON to stdout).

API routes (`internal/app/api/routes.go`):

| Method | Path                     | Notes                                             |
| ------ | ------------------------ | ------------------------------------------------- |
| POST   | `/uploads`               | JSON `{filename,content_type,size}` → presigned PUT |
| POST   | `/uploads/:id/complete`  | verifies the object, marks the row `ready`        |
| GET    | `/images?limit=&offset=` | `ready` only; limit capped at 20; presigned URLs  |
| GET    | `/image/:id`             | 404 while `pending`                               |
| DELETE | `/image/:id`             | deletes row, then object                          |

## Commands

```bash
docker compose -f compose.dev.yaml up -d      # Postgres + MinIO (needs minio.license)
make run/api                                  # API on :8080 with config.dev.yaml
cd web && VITE_API_URL=http://localhost:8080 npm run dev
cd web && npm run typecheck

make audit                                    # tidy, fmt, vet, staticcheck, go test -race
make run/cli/cleanup                          # remove stale pending uploads + objects
make build/api
go test ./...
go test -tags=integration ./internal/storage/ # against a running MinIO
docker compose -f compose.yaml up --build     # full stack behind nginx on :80

make tf/init tf/plan tf/apply tf/destroy      # MinIO storage via OpenTofu/Terraform
make run/api/tf                               # API on :8080 with config.tf.yaml
```

## Code Style

Go:

- Standard `gofmt` / `go vet` / `staticcheck` clean.
- Wrap errors with context: `fmt.Errorf("unable to ...: %v", err)`, matching existing handlers.
- Respond via `app.ErrorResponse.*` for errors and `reqres.WriteJSON` with an `envelope` for
  success. Don't write to `http.ResponseWriter` directly.
- Log with zerolog structured fields (`.Int64("image_id", id)`), not string formatting, in the
  data layer.
- Use parameterized SQL (`$1`, `$2`); never interpolate values.
- Keep `pkg/` free of `internal/` imports.

Frontend:

- TypeScript, React 19 function components with hooks, Tailwind utility classes.
- Route modules use generated `./+types/<route>` types; register new routes in `web/app/routes.ts`.

General:

- Go tests: table-driven, `_test.go` beside the code, stdlib `testing`. Handler tests use fake
  `ObjectStore` / `IImageModel`; the storage integration test is behind `-tags=integration`.
- Minimal comments — only for non-obvious logic.
