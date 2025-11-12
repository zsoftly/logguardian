# KMS key for CloudWatch Logs encryption

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_kms_key" "logguardian" {
  count = var.create_kms_key ? 1 : 0

  description             = "${var.product_name} KMS key for ${var.environment} environment"
  deletion_window_in_days = 30
  enable_key_rotation     = var.enable_key_rotation

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable root permissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow CloudWatch Logs"
        Effect = "Allow"
        Principal = {
          Service = "logs.${data.aws_region.current.id}.amazonaws.com"
        }
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:ReEncrypt*",
          "kms:GenerateDataKey*",
          "kms:DescribeKey"
        ]
        Resource = "*"
        Condition = {
          ArnEquals = {
            "kms:EncryptionContext:aws:logs:arn" = "arn:aws:logs:${data.aws_region.current.id}:${data.aws_caller_identity.current.account_id}:*"
          }
        }
      }
    ]
  })

  tags = merge(
    var.tags,
    {
      Name        = "${var.product_name}-logs-${var.environment}"
      Environment = var.environment
      ManagedBy   = "Terraform"
      Module      = "kms"
    }
  )
}

resource "aws_kms_alias" "logguardian" {
  count = var.create_kms_key ? 1 : 0

  name          = var.kms_key_alias
  target_key_id = aws_kms_key.logguardian[0].key_id
}
