variable "minio_server" {
  description = "MinIO API endpoint as host:port, without scheme."
  type        = string
  default     = "localhost:9000"
}

variable "minio_root_user" {
  description = "MinIO root user used by the provider to manage resources."
  type        = string
  default     = "minioadmin"
}

variable "minio_root_password" {
  description = "MinIO root password used by the provider to manage resources."
  type        = string
  default     = "minioadmin"
  sensitive   = true
}

variable "bucket_name" {
  description = "Name of the bucket the API reads and writes."
  type        = string
  default     = "images"
}

variable "api_user" {
  description = "IAM user the API authenticates as."
  type        = string
  default     = "img-api"
}

variable "force_destroy" {
  description = "Delete the bucket and its objects on destroy. Keep true for local only."
  type        = bool
  default     = true
}

variable "api_config_path" {
  description = "Where to write the generated API config, relative to this module."
  type        = string
  default     = "../../../config.tf.yaml"
}

variable "db_dsn" {
  description = "Postgres DSN written into the generated API config."
  type        = string
  default     = "postgres://postgres:postgres@localhost:5432/app_db?sslmode=disable"
}

variable "trusted_origins" {
  description = "CORS trusted origins written into the generated API config."
  type        = list(string)
  default     = ["http://localhost:3000", "http://localhost:5173"]
}

variable "internal_endpoint" {
  description = "Endpoint the API uses for Head/Delete, as a URL."
  type        = string
  default     = "http://localhost:9000"
}

variable "public_endpoint" {
  description = "Endpoint the browser uses, baked into presigned URLs, as a URL."
  type        = string
  default     = "http://localhost:9000"
}
