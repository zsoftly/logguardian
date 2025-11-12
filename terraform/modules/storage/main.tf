# S3 buckets for AWS Config data storage

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# S3 bucket for AWS Config
resource "aws_s3_bucket" "config" {
  count = var.create_config_bucket ? 1 : 0

  bucket = "logguardian-config-${var.environment}-${data.aws_caller_identity.current.account_id}-${data.aws_region.current.id}"

  tags = merge(
    var.tags,
    {
      Name        = "logguardian-config-${var.environment}"
      Environment = var.environment
      Purpose     = "AWS Config delivery channel storage"
      ManagedBy   = "Terraform"
      Module      = "storage"
    }
  )
}

# Block all public access
resource "aws_s3_bucket_public_access_block" "config" {
  count  = var.create_config_bucket ? 1 : 0
  bucket = aws_s3_bucket.config[0].id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Enable server-side encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "config" {
  count  = var.create_config_bucket ? 1 : 0
  bucket = aws_s3_bucket.config[0].id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Enable versioning
resource "aws_s3_bucket_versioning" "config" {
  count  = var.create_config_bucket ? 1 : 0
  bucket = aws_s3_bucket.config[0].id

  versioning_configuration {
    status = var.enable_lifecycle_rules ? "Enabled" : "Suspended"
  }
}

# Lifecycle rules for cost optimization
resource "aws_s3_bucket_lifecycle_configuration" "config" {
  count  = var.create_config_bucket && var.enable_lifecycle_rules ? 1 : 0
  bucket = aws_s3_bucket.config[0].id

  rule {
    id     = "ConfigDataExpiration"
    status = "Enabled"

    filter {}

    expiration {
      days = var.s3_expiration_days
    }

    noncurrent_version_expiration {
      noncurrent_days = 7
    }
  }
}

# Bucket policy for AWS Config service
resource "aws_s3_bucket_policy" "config" {
  count  = var.create_config_bucket ? 1 : 0
  bucket = aws_s3_bucket.config[0].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AWSConfigBucketPermissionsCheck"
        Effect = "Allow"
        Principal = {
          Service = "config.amazonaws.com"
        }
        Action   = "s3:GetBucketAcl"
        Resource = aws_s3_bucket.config[0].arn
        Condition = {
          StringEquals = {
            "AWS:SourceAccount" = data.aws_caller_identity.current.account_id
          }
        }
      },
      {
        Sid    = "AWSConfigBucketExistenceCheck"
        Effect = "Allow"
        Principal = {
          Service = "config.amazonaws.com"
        }
        Action   = "s3:ListBucket"
        Resource = aws_s3_bucket.config[0].arn
        Condition = {
          StringEquals = {
            "AWS:SourceAccount" = data.aws_caller_identity.current.account_id
          }
        }
      },
      {
        Sid    = "AWSConfigBucketPutObject"
        Effect = "Allow"
        Principal = {
          Service = "config.amazonaws.com"
        }
        Action   = "s3:PutObject"
        Resource = "${aws_s3_bucket.config[0].arn}/*"
        Condition = {
          StringEquals = {
            "s3:x-amz-acl"      = "bucket-owner-full-control"
            "AWS:SourceAccount" = data.aws_caller_identity.current.account_id
          }
        }
      }
    ]
  })
}
