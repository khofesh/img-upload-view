# PRD — Object Storage with Presigned Uploads

- Status: Draft
- Date: 2026-09-28
- Plan: [object-storage-presigned-upload-plan.md](./object-storage-presigned-upload-plan.md)

## 1. Background

Today the browser posts the JPEG as `multipart/form-data` to `POST /upload`. The Go API holds the
whole body (up to 10 MB), writes it to local disk (`UPLOAD_DIR`), and stores `/images/<filename>`
in Postgres. Images are served by nginx from a shared Docker volume (prod) or by the API itself
(`env: local`).

Problems:

- Every upload passes through the API process, so API memory, bandwidth and `WriteTimeout` (10 s)
  limit uploads.
- File storage is tied to one host's disk and a shared Docker volume, so the API can't scale
  horizontally.
- Disk and DB can drift apart: the disk write and the DB insert aren't atomic, and deletes can
  leave orphan files.
- `UPLOAD_DIR` has two different defaults (`/app/uploads` and `./upload`), which already causes
  confusion.

## 2. Goal

Store images in S3-compatible object storage. The browser uploads straight to the storage using
a short-lived presigned URL issued by the API. The API only handles metadata and signatures.

## 3. Non-goals

- Authentication / per-user ownership of images.
- Image processing (thumbnails, resizing, EXIF stripping).
- Multipart/resumable uploads (files are ≤ 10 MB; a single PUT is enough).
- Supporting formats other than JPEG.
- A CDN in front of the bucket.

## 4. Storage backend decision

**Recommendation: MinIO for local dev and the compose "fake prod". The Go code talks plain S3 API
only (`aws-sdk-go-v2`), so LocalStack or real AWS S3 are config-only swaps.**

|                          | MinIO                                                                                                               | LocalStack                                      |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------- |
| What it is               | S3-compatible object server; can run in real deployments                                                            | AWS emulator for dev/test                       |
| Fits "fake prod" compose | Yes, a real storage server                                                                                          | Emulator only                                   |
| Bucket CORS              | Server-level (`MINIO_API_CORS_ALLOW_ORIGIN`); bucket CORS API support is limited                                    | `PutBucketCors` supported                       |
| Footprint                | Single small binary                                                                                                 | Heavier, emulates many services                 |
| Risk                     | Community edition distribution/licensing changed in 2025. Pin an image tag and check it's available before starting | Check current licensing/auth-token requirements |

Because the code only uses the S3 API through configurable endpoints, the choice is limited to
`compose*.yaml` and config files.

## 5. User stories

1. As a user, I pick a JPEG ≤ 10 MB and it uploads, with the same UX as today.
2. As a user, I see a clear error before any upload starts if the file is too large or not a JPEG.
3. As a user, the gallery shows my uploaded images and I can delete them.
4. As a developer, `docker compose -f compose.dev.yaml up -d` gives me Postgres and object storage
   with the bucket ready, with no manual setup.
5. As an operator, no image bytes pass through the API.

## 6. Functional requirements

### Upload flow

```
Browser                         API                          Object storage        Postgres
  │ POST /uploads                │                                 │                   │
  │ {filename,content_type,size} │                                 │                   │
  │─────────────────────────────▶│ validate (jpeg, ≤10MB)          │                   │
  │                              │ key = images/<unix>_<hex>.jpg   │                   │
  │                              │ INSERT status='pending' ───────────────────────────▶│
  │                              │ presign PUT (TTL 5m)            │                   │
  │◀─────────────────────────────│ {id, upload:{url,method,headers,expires_at}}        │
  │ PUT <presigned url> (bytes) ──────────────────────────────────▶│                   │
  │◀─────────────────────────────────────────────────────── 200 ───│                   │
  │ POST /uploads/:id/complete ─▶│ HeadObject(key) ───────────────▶│                   │
  │                              │ verify exists, size, type       │                   │
  │                              │ UPDATE status='ready' ─────────────────────────────▶│
  │◀─────────────────────────────│ 200 {image}                     │                   │
```

- **FR1** `POST /uploads` takes JSON `{filename, content_type, size}`. It rejects anything other
  than `image/jpeg` / `image/jpg`, `size <= 0`, and `size > 10 MB` with the existing
  `BadRequestResponse`.
- **FR2** The server generates the object key. Clients never choose keys, so they can't overwrite
  objects.
- **FR3** The presigned PUT signs `Content-Type` and `Content-Length`, so the storage rejects a body
  that differs from what was declared. It expires in 5 minutes (configurable).
- **FR4** `POST /uploads/:id/complete` checks the object with `HeadObject`:
  - It returns 422 when the object is missing, larger than 10 MB, or doesn't match the declared
    size/type. When the object exists but is invalid, the server also deletes it.
  - On success it sets the row to `ready`.
  - Calling it again on a `ready` row returns 200 with the image (idempotent).
- **FR5** Optional hardening: before marking `ready`, check the JPEG magic bytes (`FF D8 FF`) with a
  ranged GET of the first 3 bytes.

### Read / delete

- **FR6** The bucket is private. `GET /images` and `GET /image/:id` return only `ready` rows, and
  each `url` is a presigned GET URL (TTL 15 min, configurable). The response shape stays the same,
  so the gallery keeps working.
- **FR7** `DELETE /image/:id` deletes the object, then the row. A missing object is not an error.
- **FR8** Pending rows older than 1 hour, and their objects if present, are removed by
  `cmd/cli cleanup` (run manually or on a cron).

### Removal of the old path

- **FR9** After the frontend switches over, the following are removed:
  - `POST /upload`
  - the local-disk writes
  - `UPLOAD_DIR`
  - the `/images/*` static file server
  - nginx `location /images/`
  - the `uploaded_images` volume

## 7. Non-functional requirements

- Upload size limit enforced at three layers: frontend check, API validation, and the signed
  `Content-Length` plus the `HeadObject` check.
- The presign endpoint makes no calls to object storage (signing is local), so p95 latency stays
  under 50 ms.
- The API must not need network access to storage to start, unless `storage.createBucket: true`.
- The same config shape works for MinIO, LocalStack and AWS S3.
- Every storage call takes `r.Context()` so requests can be cancelled.

## 8. API contract

`POST /uploads`

```json
// request
{ "filename": "cat.jpg", "content_type": "image/jpeg", "size": 482113 }
// 201
{
  "id": 42,
  "upload": {
    "url": "http://localhost:9000/images/images/1759050000_ab12cd34ef56ab78.jpg?X-Amz-...",
    "method": "PUT",
    "headers": { "Content-Type": "image/jpeg" },
    "expires_at": "2026-09-28T10:05:00Z"
  }
}
```

`POST /uploads/:id/complete` → `200 {"message": "...", "image": {...}}` · `404` unknown id ·
`422` object missing/invalid.

`GET /images`, `GET /image/:id`: unchanged shape. `image.url` is now an absolute presigned GET URL.

## 9. Data model

Changes to the `images` table:

| Column                                          | Change                                                           |
| ----------------------------------------------- | ---------------------------------------------------------------- |
| `object_key VARCHAR(500) NOT NULL`              | new, unique                                                      |
| `status VARCHAR(16) NOT NULL DEFAULT 'pending'` | new, `pending` or `ready`                                        |
| `url`                                           | dropped. URLs are generated per request                          |
| `file_size`                                     | stores the declared size, then the size verified by `HeadObject` |

The existing `upload_timestamp` index is replaced with a partial index
`WHERE status = 'ready'`.

## 10. Success criteria

- Upload, view and delete work end-to-end on `compose.dev.yaml` (MinIO) and on the `compose.yaml`
  full stack.
- Uploading an 11 MB file is rejected by the frontend, and by the API/storage when the frontend is
  bypassed with curl.
- No image bytes reach the API: `POST /uploads` requests are small JSON only.
- `make audit` passes. New handlers and the storage adapter have unit tests.

## 11. Risks & open questions

| #   | Item                                                                                                     | Proposed answer                                                                                 |
| --- | -------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| Q1  | MinIO or LocalStack?                                                                                     | MinIO (section 4). Revisit if the pinned MinIO image isn't available                            |
| Q2  | Migrate existing files in `upload/` or the volume?                                                       | No, drop them. This is pre-production. A `cli migrate-local` command is an optional follow-up   |
| Q3  | Presigned GET or public-read bucket for viewing?                                                         | Presigned GET: keeps the bucket private. The downside is that URLs change and caching is weaker |
| Q4  | Host mismatch: the API reaches `minio:9000` but browsers need `localhost:9000`, and SigV4 signs the host | Two clients: an internal endpoint for Head/Delete and a `publicEndpoint` for presigning         |
| Q5  | Does `aws-sdk-go-v2` presign actually enforce `Content-Length` against MinIO?                            | Test it in phase 1. Fallback: the `HeadObject` check in FR4 still enforces the limit            |
| Q6  | Browser CORS for PUT to storage                                                                          | MinIO: `MINIO_API_CORS_ALLOW_ORIGIN`. LocalStack/S3: `PutBucketCors` at bucket creation         |
| Q7  | No auth, so anyone can request presigned URLs                                                            | Same exposure as today's open `/upload`. Rate limiting is out of scope                          |
