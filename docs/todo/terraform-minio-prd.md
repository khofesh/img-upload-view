# PRD: Terraform-managed MinIO storage (local)

## Problem

Object storage setup is implicit and done by the app itself:

- `cmd/api/main.go` calls `store.EnsureBucket()` when `storage.createBucket: true`, so the API
  creates the `images` bucket on startup.
- `S3Store.ensureCORS` tries `PutBucketCors`, which MinIO rejects (`NotImplemented`); CORS actually
  comes from `MINIO_API_CORS_ALLOW_ORIGIN` in compose.
- The API authenticates with the MinIO root user (`minioadmin` / `minioadmin`), which is also
  committed in `config.dev.yaml` and `configs/config.yaml`.

As a result, the API has full admin rights on storage. Nothing records what the bucket should look
like, and local setup is nothing like what a real deployment would provision.

## Goal

Declare the storage resources the app depends on in Terraform and apply them against the local
MinIO started by `compose.dev.yaml`. The API should then run with least-privilege credentials and
no longer create the bucket itself.

## Scope

In scope:

- Terraform (OpenTofu-compatible) root module under `infra/terraform/minio/` using the
  `aminueza/minio` provider.
- Bucket `images`: private, no versioning, `force_destroy` only in local.
- IAM policy scoped to `arn:aws:s3:::images/*` allowing `s3:PutObject`, `s3:GetObject`,
  `s3:DeleteObject`, plus `s3:ListBucket` / `s3:GetBucketLocation` on the bucket.
  (`Head` = `GetObject`; presigned PUT/GET are signed with these credentials.)
- IAM user `img-api` with that policy attached; its access key and secret are exported as
  sensitive outputs.
- A generated, gitignored API config (`config.tf.yaml`) that uses the `img-api` credentials and
  sets `createBucket: false`.
- Make targets: `tf/init`, `tf/plan`, `tf/apply`, `tf/destroy`, `run/api/tf`.
- Local state (`terraform.tfstate`, gitignored).

Out of scope:

- Containers, networks and volumes. `docker compose` keeps managing them.
- PostgreSQL roles and schema. `01-init.sql` stays as is.
- Remote state and cloud providers (AWS, LocalStack).
- The prod `compose.yaml` stack. This could follow later with a second tfvars file.
- CORS. MinIO only supports it through server env, so it stays in compose.

## Requirements

| #   | Requirement                                                                                                                                           |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| R1  | Running `make tf/apply` on a fresh `compose.dev.yaml` stack creates the bucket, policy and user, and is idempotent: a second `plan` shows no changes. |
| R2  | The API started with the generated config can upload, complete, list, get and delete images end to end.                                               |
| R3  | The `img-api` credentials cannot create or delete buckets or reach other buckets, which is verified with `mc`.                                        |
| R4  | No secret is committed: state, generated config and `*.tfvars` holding secrets are gitignored.                                                        |
| R5  | The existing flow (`make run/api` with `config.dev.yaml` and `createBucket: true`) keeps working unchanged for anyone not using Terraform.            |
| R6  | `tofu` and `terraform` both work; the pinned version is `>= 1.6`.                                                                                     |

## Success criteria

- A fresh clone followed by `compose up`, `make tf/init tf/apply` and `make run/api/tf` gives a
  working upload in the browser.
- `mc admin user info` shows `img-api` with only the `img-api-images-rw` policy.
- `make audit` and `go test ./...` still pass, and `go test -tags=integration ./internal/storage/`
  passes when run with the Terraform credentials.

## Open questions

1. Should `config.dev.yaml` switch to Terraform credentials by default later, making Terraform
   required for local dev? For now it's opt-in.
2. Should Terraform also manage a lifecycle rule that expires orphaned objects under `images/`? It
   would overlap with `cmd/cli cleanup`, which is DB-driven, so it's deferred.
