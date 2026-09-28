# AGENTS.md

## Project

Image upload & gallery service. Go 1.24 JSON API + React Router 7 (SSR) frontend, PostgreSQL
for metadata, uploaded JPEGs stored on local disk.

```
cmd/api/                  API entry point (flag -config-path, default /etc/secrets/config.yaml)
cmd/cli/                  CLI stub
internal/app/api/         server.go (http.Server + graceful shutdown), routes.go (httprouter)
internal/app/api/handlers/ image.go — upload / list / get / delete handlers
internal/config/          Config (YAML) + Application (DI container passed to handlers)
internal/data/            Models, IImageModel interface, ImageModel (database/sql + lib/pq)
internal/db/              OpenDB — pool setup + ping
internal/middleware/      Generic Middlewares[T] with functional options: CORS, RecoverPanic
internal/reqres/          Request param readers, WriteJSON, WriteFile
pkg/errors/               ErrorResponse — JSON error envelope helpers
pkg/read-config/          YAML config loader
configs/                  prod config.yaml, nginx.conf, postgres-init/01-init.sql (schema)
web/                      React Router 7 + Vite + Tailwind 4 frontend (routes: home, upload, gallery)
dummy-jpeg/               Python script generating oversized test JPEGs
```

Key facts:

- No migrations tool. Schema lives in `configs/postgres-init/01-init.sql`, applied only on first
  Postgres container init. Schema changes require editing that file and recreating the volume.
- Handlers are closures: `func X(app *config.Application) http.HandlerFunc`. Dependencies go
  through `config.Application` (`Logger`, `Config`, `Models`, `ErrorResponse`).
- Data access goes through `data.Models.Image` (`IImageModel`). Add new queries to the interface
  and `ImageModel`.
- "Not found" is signalled by `errors.New("record not found")` and matched by string in
  handlers — keep that message stable.
- Uploads: form field `image`, JPEG only, max 10 MB (`MaxUploadSize`). Stored as
  `<unix>_<16 hex>.<ext>` in `UPLOAD_DIR`; DB row stores URL `/images/<filename>`.
- `UPLOAD_DIR` defaults differ: handlers fall back to `/app/uploads`, the local static file
  server to `./upload`. Set `UPLOAD_DIR=./upload/` when running locally.
- In `env: local` the API serves `/images/*` itself; in prod nginx serves them from the shared
  `uploaded_images` volume and proxies `/api/*` → API (prefix stripped) and `/` → frontend.
- Config is YAML only (`config.dev.yaml` locally, `configs/config.yaml` in prod). The `PORT` /
  `DB_DSN` env vars in `compose.yaml` are not read by the code.
- Frontend calls `${VITE_API_URL || "/api"}`. Dev: `VITE_API_URL=http://localhost:8080`.
- Logging: zerolog (global `log.Logger`, JSON to stdout).

API routes (`internal/app/api/routes.go`):

| Method | Path                     | Notes                                             |
| ------ | ------------------------ | ------------------------------------------------- |
| POST   | `/upload`                | multipart, field `image`                          |
| GET    | `/images?limit=&offset=` | limit capped at 20; returns `images` + `metadata` |
| GET    | `/image/:id`             |                                                   |
| DELETE | `/image/:id`             | deletes row, then file                            |

## Commands

```bash
docker compose -f compose.dev.yaml up -d      # Postgres only
UPLOAD_DIR=./upload/ make run/api             # API on :8080 with config.dev.yaml
cd web && VITE_API_URL=http://localhost:8080 npm run dev
cd web && npm run typecheck

make audit                                    # tidy, fmt, vet, staticcheck, go test -race
make build/api
go test ./...
docker compose -f compose.yaml up --build     # full stack behind nginx on :80
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

- No tests exist yet. New Go tests: table-driven, `_test.go` beside the code, stdlib `testing`.
- Minimal comments — only for non-obvious logic.
