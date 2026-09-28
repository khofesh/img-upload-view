output "bucket" {
  description = "Name of the managed bucket."
  value       = minio_s3_bucket.images.bucket
}

output "api_access_key" {
  description = "Access key of the API IAM user."
  value       = minio_iam_user.api.id
}

output "api_secret_key" {
  description = "Secret key of the API IAM user."
  value       = minio_iam_user.api.secret
  sensitive   = true
}
