# Terraform: MinIO storage (local)

Declares the `images` bucket, the least-privilege `img-api` user and its policy against the MinIO
started by `compose.dev.yaml`, then writes a gitignored API config (`config.tf.yaml`) that uses
those credentials with `createBucket: false`.

```bash
docker compose -f compose.dev.yaml up -d
cp local.tfvars.example local.tfvars
make tf/init tf/apply
make run/api/tf
```

If the `images` bucket already exists (for example because the API started once with
`createBucket: true`), `apply` fails with "bucket already exists". Either adopt it with
`$(TF) -chdir=infra/terraform/minio import minio_s3_bucket.images images`, or recreate the stack
with `docker compose -f compose.dev.yaml down -v` first.

The `img-api` secret is stored in plaintext in local state. That is fine for local development;
remote state would need encryption. `config.tf.yaml` is written with mode `0600` and is gitignored.
