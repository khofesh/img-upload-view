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
