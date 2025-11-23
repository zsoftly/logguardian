# ============================================
# KMS Key for CloudWatch Logs Encryption
# ============================================

resource "aws_kms_key" "logs" {
  count = var.create_kms_key ? 1 : 0

  description             = "KMS key for CloudWatch Logs encryption (${var.environment})"
  deletion_window_in_days = var.kms_deletion_window_days
  enable_key_rotation     = true

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:${data.aws_partition.current.partition}:iam::${local.account_id}:root"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow CloudWatch Logs"
        Effect = "Allow"
        Principal = {
          Service = "logs.${local.region}.amazonaws.com"
        }
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:ReEncrypt*",
          "kms:GenerateDataKey*",
          "kms:CreateGrant",
          "kms:DescribeKey"
        ]
        Resource = "*"
        Condition = {
          ArnLike = {
            "kms:EncryptionContext:aws:logs:arn" = "arn:${data.aws_partition.current.partition}:logs:${local.region}:${local.account_id}:log-group:*"
          }
        }
      },
      {
        Sid    = "Allow Lambda Access"
        Effect = "Allow"
        Principal = {
          AWS = local.lambda_role_arn
        }
        Action = [
          "kms:DescribeKey",
          "kms:GetKeyPolicy",
          "kms:Decrypt"
        ]
        Resource = "*"
      }
    ]
  })

  tags = merge(
    local.common_tags,
    {
      Name = local.kms_key_alias
    }
  )
}

resource "aws_kms_alias" "logs" {
  count = var.create_kms_key ? 1 : 0

  name          = "alias/${local.kms_key_alias}"
  target_key_id = aws_kms_key.logs[0].key_id
}
