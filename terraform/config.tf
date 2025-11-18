# ============================================
# AWS Config Service (Optional)
# ============================================

resource "aws_s3_bucket" "config" {
  count = var.create_config_service ? 1 : 0

  bucket        = local.config_bucket_name
  force_destroy = false

  tags = merge(
    local.common_tags,
    {
      Name = local.config_bucket_name
    }
  )
}

resource "aws_s3_bucket_lifecycle_configuration" "config" {
  count = var.create_config_service ? 1 : 0

  bucket = aws_s3_bucket.config[0].bucket

  rule {
    id     = "expire-old-config-snapshots"
    status = "Enabled"

    filter {
      prefix = ""
    }

    expiration {
      days = var.config_bucket_expiration_days
    }

    noncurrent_version_expiration {
      noncurrent_days = 30
    }

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}

resource "aws_s3_bucket_versioning" "config" {
  count = var.create_config_service ? 1 : 0

  bucket = aws_s3_bucket.config[0].bucket

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "config" {
  count = var.create_config_service ? 1 : 0

  bucket = aws_s3_bucket.config[0].id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "config" {
  count = var.create_config_service ? 1 : 0

  bucket = aws_s3_bucket.config[0].id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# ============================================
# AWS Config Recorder
# ============================================

resource "aws_config_configuration_recorder" "main" {
  count = var.create_config_service ? 1 : 0

  name     = "${local.name_prefix}-recorder"
  role_arn = local.config_role_arn

  recording_group {
    all_supported                 = false
    include_global_resource_types = false
    resource_types                = var.config_recorder_resource_types
  }
}

resource "aws_config_delivery_channel" "main" {
  count = var.create_config_service ? 1 : 0

  name           = "${local.name_prefix}-delivery-channel"
  s3_bucket_name = aws_s3_bucket.config[0].bucket

  depends_on = [aws_config_configuration_recorder.main]
}

resource "aws_config_configuration_recorder_status" "main" {
  count = var.create_config_service ? 1 : 0

  name       = aws_config_configuration_recorder.main[0].name
  is_enabled = true

  depends_on = [aws_config_delivery_channel.main]
}

# ============================================
# Config Rule - CloudWatch Log Group Encryption
# ============================================

resource "aws_config_config_rule" "encryption" {
  count = var.create_config_rules ? 1 : 0

  name        = local.encryption_rule_name
  description = "Checks if CloudWatch Log Groups are encrypted with KMS"

  source {
    owner             = "AWS"
    source_identifier = "CLOUDWATCH_LOG_GROUP_ENCRYPTED"
  }

  input_parameters = jsonencode({
    KmsKeyId = local.kms_key_arn
  })

  depends_on = [
    aws_config_configuration_recorder.main,
    aws_config_delivery_channel.main
  ]

  tags = merge(
    local.common_tags,
    {
      Name = local.encryption_rule_name
    }
  )
}

# ============================================
# Config Rule - CloudWatch Log Group Retention
# ============================================

resource "aws_config_config_rule" "retention" {
  count = var.create_config_rules ? 1 : 0

  name        = local.retention_rule_name
  description = "Checks if CloudWatch Log Groups have retention policies set"

  source {
    owner             = "AWS"
    source_identifier = "CW_LOGGROUP_RETENTION_PERIOD_CHECK"
  }

  input_parameters = jsonencode({
    MinRetentionTime = var.default_retention_days
    MaxRetentionTime = var.default_retention_days
  })

  depends_on = [
    aws_config_configuration_recorder.main,
    aws_config_delivery_channel.main
  ]

  tags = merge(
    local.common_tags,
    {
      Name = local.retention_rule_name
    }
  )
}

# ============================================
# Config Remediation Configuration
# ============================================
# Remediation uses SNS notification for human review workflow.
# Automatic remediation is disabled to prevent unintended changes to production log groups.
# Operations team reviews notifications and manually approves remediation.
resource "aws_config_remediation_configuration" "encryption" {
  count = var.create_config_rules && var.alarm_sns_topic_arn != null ? 1 : 0

  config_rule_name = aws_config_config_rule.encryption[0].name
  target_type      = "SSM_DOCUMENT"
  target_id        = "AWS-PublishSNSNotification"
  automatic        = false

  parameter {
    name         = "AutomationAssumeRole"
    static_value = local.config_role_arn
  }

  parameter {
    name         = "TopicArn"
    static_value = var.alarm_sns_topic_arn
  }

  parameter {
    name         = "Message"
    static_value = "AWS Config Alert: Non-compliant CloudWatch Logs encryption detected."
  }

  depends_on = [aws_config_config_rule.encryption]
}

resource "aws_config_remediation_configuration" "retention" {
  count = var.create_config_rules ? 1 : 0

  config_rule_name = aws_config_config_rule.retention[0].name
  target_type      = "SSM_DOCUMENT"
  target_id        = "AWS-PublishSNSNotification"
  automatic        = false

  parameter {
    name         = "AutomationAssumeRole"
    static_value = local.lambda_role_arn
  }

  parameter {
    name           = "Message"
    resource_value = "RESOURCE_ID"
  }
}