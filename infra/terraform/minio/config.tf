resource "local_sensitive_file" "api_config" {
  filename        = var.api_config_path
  file_permission = "0600"
  content = templatefile("${path.module}/templates/config.yaml.tftpl", {
    db_dsn            = var.db_dsn
    trusted_origins   = var.trusted_origins
    internal_endpoint = var.internal_endpoint
    public_endpoint   = var.public_endpoint
    bucket            = minio_s3_bucket.images.bucket
    access_key        = minio_iam_user.api.id
    secret_key        = minio_iam_user.api.secret
  })
}
