terraform {
  required_version = ">= 1.6"

  required_providers {
    minio = {
      source  = "aminueza/minio"
      version = "~> 3.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.0"
    }
  }
}
