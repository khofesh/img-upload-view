# Plan: Terraform-managed MinIO storage (local)

See `terraform-minio-prd.md`.

## Layout

```
infra/terraform/minio/
  versions.tf          terraform >= 1.6, aminueza/minio ~> 3.x
  providers.tf         provider "minio" (server, user, password from vars)
  variables.tf         minio_server, minio_root_user, minio_root_password (sensitive),
                       bucket_name = "images", api_user = "img-api", force_destroy = true,
                       api_config_path, db_dsn, trusted_origins, public_endpoint, internal_endpoint
  main.tf              bucket, policy, user, attachment
  config.tf            local_sensitive_file → generated API config
  templates/config.yaml.tftpl
  outputs.tf           bucket, api_access_key, api_secret_key (sensitive)
  local.tfvars.example
```

## Steps

### 1. Scaffold module

- `versions.tf`: `required_providers { minio = { source = "aminueza/minio" } local = {...} }`.
- `providers.tf`:
  ```hcl
  provider "minio" {
    minio_server   = var.minio_server        # "localhost:9000"
    minio_user     = var.minio_root_user
    minio_password = var.minio_root_password
    minio_ssl      = false
  }
  ```
- `local.tfvars.example` with the compose dev values; `local.tfvars` is gitignored.

### 2. Resources (`main.tf`)

```hcl
resource "minio_s3_bucket" "images" {
  bucket        = var.bucket_name
  acl           = "private"
  force_destroy = var.force_destroy
}

data "minio_iam_policy_document" "api" {
  statement {
    actions   = ["s3:GetBucketLocation", "s3:ListBucket"]
    resources = ["arn:aws:s3:::${minio_s3_bucket.images.bucket}"]
  }
  statement {
    actions   = ["s3:PutObject", "s3:GetObject", "s3:DeleteObject"]
    resources = ["arn:aws:s3:::${minio_s3_bucket.images.bucket}/*"]
  }
}

resource "minio_iam_policy" "api" {
  name   = "${var.api_user}-${var.bucket_name}-rw"
  policy = data.minio_iam_policy_document.api.json
}

resource "minio_iam_user" "api" {
  name = var.api_user
}

resource "minio_iam_user_policy_attachment" "api" {
  user_name   = minio_iam_user.api.id
  policy_name = minio_iam_policy.api.id
}
```

The access key is the user name and the secret is `minio_iam_user.api.secret`. If the provider
version prefers it, switch to `minio_iam_service_account`; the outputs don't change.

### 3. Generated API config (`config.tf`)

- Template mirrors `config.dev.yaml` with:
  `accessKey = var.api_user`, `secretKey = minio_iam_user.api.secret`,
  `bucket = minio_s3_bucket.images.bucket`, `createBucket: false`.
- `local_sensitive_file` writes to `var.api_config_path` (default `../../../config.tf.yaml`),
  `file_permission = "0600"`.

### 4. Repo wiring

- `.gitignore`: `infra/terraform/**/.terraform/`, `*.tfstate*`, `*.tfvars` (keep
  `*.tfvars.example`), `config.tf.yaml`. Commit `.terraform.lock.hcl`.
- `Makefile`:
  ```make
  TF ?= $(shell command -v tofu || command -v terraform)
  TF_DIR = infra/terraform/minio
  tf/init:    ; $(TF) -chdir=$(TF_DIR) init
  tf/plan:    ; $(TF) -chdir=$(TF_DIR) plan -var-file=local.tfvars
  tf/apply:   ; $(TF) -chdir=$(TF_DIR) apply -var-file=local.tfvars
  tf/destroy: ; $(TF) -chdir=$(TF_DIR) destroy -var-file=local.tfvars
  run/api/tf: ; go run ./cmd/api -config-path=./config.tf.yaml
  ```
- `AGENTS.md`: add `infra/terraform/minio/` to the layout and the `tf/*` targets to Commands.

### 5. Adoption of an existing bucket

If `images` already exists from `EnsureBucket`, `apply` fails with "bucket already exists". To fix
it, either run `$(TF) import minio_s3_bucket.images images` or run
`docker compose -f compose.dev.yaml down -v` first. Document this in the README of the module
directory. It's one paragraph, not a separate file.

### 6. Verification

1. `docker compose -f compose.dev.yaml down -v && docker compose -f compose.dev.yaml up -d`
2. `make tf/init tf/apply`; then `make tf/plan` shows `No changes` (R1).
3. `make run/api/tf` and the web dev server: upload a JPEG, view it in the gallery, delete it (R2).
4. `mc alias set tf http://localhost:9000 img-api <secret>`:
   - `mc mb tf/other` is denied, `mc rb tf/images` is denied, and `mc cp x.jpg tf/images/x.jpg`
     succeeds (R3).
5. `git status` shows no state, tfvars or generated config (R4).
6. `make run/api` with `config.dev.yaml` still works (R5).
7. Repeat step 2 with the other binary (`TF=terraform` / `TF=tofu`) (R6).
8. `make audit`.

## Risks

- **Provider drift on MinIO AIStor**: `aminueza/minio` targets upstream MinIO. Check the IAM and
  bucket resources against the AIStor image in step 6 before building on them.
- **Secret in state**: `minio_iam_user.secret` is stored in plaintext in local state. That's
  acceptable locally; remote state would need encryption (OpenTofu state encryption or a backend
  that encrypts it).
- **`PutBucketCors` warning**: the warning logged on every start disappears because
  `createBucket: false` skips `EnsureBucket`. Leave the code as is.

## Follow-ups (not in this change)

- A `prod.tfvars` targeting the `compose.yaml` MinIO, with `force_destroy = false`.
- Remove root credentials from `configs/config.yaml` once prod uses Terraform.
- Optional lifecycle rule (see PRD open question 2).
