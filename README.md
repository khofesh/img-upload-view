# update load and view image services

`compose.dev.yaml` is for dev

`compose.yaml` is for prod (let's say it's prod)

Images are stored in S3-compatible object storage (MinIO AIStor). The API only signs
presigned URLs; image bytes go straight from the browser to storage.

## selinux

```shell
sudo chgrp -R nogroup configs
sudo chcon -Rt svirt_sandbox_file_t configs/
```

## MinIO license

MinIO AIStor requires a license file. Put it in the repo root as `minio.license` (it is
gitignored), or point `MINIO_LICENSE_PATH` at it:

```shell
export MINIO_LICENSE_PATH="$HOME/minio.license"
```

## development

docker (Postgres + MinIO) and API service (terminal 1)

```shell
# docker compose
docker compose -f compose.dev.yaml up -d

make run/api
```

frontend (terminal 2)

```shell
export VITE_API_URL=http://localhost:8080
cd web
npm run dev
```

MinIO console: <http://localhost:9001> (user/password `minioadmin` / `minioadmin`).

requests

The upload is a three-step flow: ask for a presigned URL, PUT the bytes to storage, then
confirm the upload.

```shell
# 1. create the upload
curl -X POST http://localhost:8080/uploads \
  -H "Content-Type: application/json" \
  -d '{"filename":"cat.jpg","content_type":"image/jpeg","size":482113}'
# => {"id":42,"upload":{"url":"...","method":"PUT","headers":{...},"expires_at":"..."}}

# 2. upload the bytes directly to storage
curl -X PUT "<upload.url>" \
  -H "Content-Type: image/jpeg" \
  --data-binary @/path/to/your/image.jpg

# 3. confirm the upload
curl -X POST http://localhost:8080/uploads/42/complete

# get all images
curl -X GET http://localhost:8080/images

# with limit and offset
curl -X GET "http://localhost:8080/images?limit=5&offset=0"

# next page
curl -X GET "http://localhost:8080/images?limit=5&offset=5"

# get image by ID
curl -X GET http://localhost:8080/image/1

# delete image by ID
curl -X DELETE http://localhost:8080/image/1
```

cleanup stale pending uploads (older than 1h by default)

```shell
make run/cli/cleanup
```

psql

```shell
psql "postgres://postgres:postgres@localhost:5432/app_db?sslmode=disable"
```

## terraform storage (optional)

By default the API runs with MinIO root credentials and creates the `images` bucket itself
(`createBucket: true` in `config.dev.yaml`). Terraform is an opt-in alternative that provisions
the bucket plus a least-privilege `img-api` IAM user, and writes a gitignored `config.tf.yaml`
with `createBucket: false` for the API to use instead.

```shell
docker compose -f compose.dev.yaml up -d
cp infra/terraform/minio/local.tfvars.example infra/terraform/minio/local.tfvars

make tf/init
make tf/apply     # idempotent; `make tf/plan` shows No changes afterwards
make run/api/tf   # API with the generated config.tf.yaml
```

`tofu` and `terraform` both work; the Makefile picks whichever is on `PATH`
(override with `make TF=terraform ...`). Tear the resources down with `make tf/destroy`.

If the `images` bucket already exists (for example from a previous `make run/api`), `apply` fails
with "bucket already exists". Adopt it with:

```shell
make tf/init
$(command -v tofu || command -v terraform) -chdir=infra/terraform/minio import minio_s3_bucket.images images
make tf/apply
```

or rebuild the stack with `docker compose -f compose.dev.yaml down -v` first. The generated
`config.tf.yaml`, local state and `local.tfvars` are all gitignored.

## fake prod

```shell
docker compose -f compose.yaml up --build # if "localhost" cannot be accessed, wait a bit
docker compose -f compose.yaml up -d
```

## generate dummy JPEG

```shell
cd dummy-jpeg
python3 gen-dummy-jpeg.py
```

test it on the webpage

![error-more-than-10mb](./images/error-more-than-10mb.png)
