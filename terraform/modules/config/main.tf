# AWS Config service and compliance rules

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# Config service role
resource "aws_iam_role" "config" {
  count = var.create_config_service ? 1 : 0

  name = "LogGuardian-ConfigRole-${var.environment}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "config.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })

  tags = merge(
    var.tags,
    {
      Name        = "LogGuardian-ConfigRole-${var.environment}"
      Environment = var.environment
      Purpose     = "Config service role"
      ManagedBy   = "Terraform"
      Module      = "config"
    }
  )
}

# Attach AWS managed policy for Config
resource "aws_iam_role_policy_attachment" "config" {
  count = var.create_config_service ? 1 : 0

  role       = aws_iam_role.config[0].name
  policy_arn = "arn:aws:iam::aws:policy/service-role/ConfigRole"
}

# Configuration recorder
resource "aws_config_configuration_recorder" "main" {
  count = var.create_config_service ? 1 : 0

  name     = "LogGuardian-Recorder-${var.environment}"
  role_arn = aws_iam_role.config[0].arn

  recording_group {
    all_supported                 = false
    include_global_resource_types = false
    resource_types                = ["AWS::Logs::LogGroup"]
  }

  depends_on = [aws_iam_role_policy_attachment.config]
}

# Delivery channel
resource "aws_config_delivery_channel" "main" {
  count = var.create_config_service ? 1 : 0

  name           = "LogGuardian-DeliveryChannel-${var.environment}"
  s3_bucket_name = var.config_bucket_name

  depends_on = [aws_config_configuration_recorder.main]
}

# Start the recorder
resource "aws_config_configuration_recorder_status" "main" {
  count = var.create_config_service ? 1 : 0

  name       = aws_config_configuration_recorder.main[0].name
  is_enabled = true

  depends_on = [aws_config_delivery_channel.main]
}

# Encryption Config rule
resource "aws_config_config_rule" "encryption" {
  count = var.create_encryption_rule ? 1 : 0

  name        = "logguardian-encryption-${var.environment}"
  description = "Checks if CloudWatch log groups are encrypted with KMS"

  source {
    owner             = "AWS"
    source_identifier = "CLOUDWATCH_LOG_GROUP_ENCRYPTED"
  }

  scope {
    compliance_resource_types = ["AWS::Logs::LogGroup"]
  }

  maximum_execution_frequency = "TwentyFour_Hours"

  depends_on = [
    aws_config_configuration_recorder.main,
    aws_config_configuration_recorder_status.main
  ]
}

# Retention Config rule
resource "aws_config_config_rule" "retention" {
  count = var.create_retention_rule ? 1 : 0

  name        = "logguardian-retention-${var.environment}"
  description = "Checks if CloudWatch log groups have retention policy set to at least ${var.default_retention_days} days"

  source {
    owner             = "AWS"
    source_identifier = "CW_LOGGROUP_RETENTION_PERIOD_CHECK"
  }

  input_parameters = jsonencode({
    MinRetentionTime = tostring(var.default_retention_days)
  })

  scope {
    compliance_resource_types = ["AWS::Logs::LogGroup"]
  }

  maximum_execution_frequency = "TwentyFour_Hours"

  depends_on = [
    aws_config_configuration_recorder.main,
    aws_config_configuration_recorder_status.main
  ]
}
