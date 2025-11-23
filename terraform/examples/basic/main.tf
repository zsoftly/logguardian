# LogGuardian - Basic Example
#
# This example demonstrates a minimal deployment of LogGuardian
# with default settings. Suitable for getting started quickly.

terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "ca-central-1"
}

module "logguardian" {
  source = "../../"

  # Required variables
  environment      = "dev"
  lambda_s3_bucket = "my-deployment-bucket"
  lambda_s3_key    = "logguardian-compliance.zip"

  # Optional: Customize owner
  owner = "Platform-Engineering"
}

# Outputs
output "lambda_function_name" {
  description = "Name of the deployed Lambda function"
  value       = module.logguardian.lambda_function_name
}

output "dashboard_url" {
  description = "URL to the CloudWatch dashboard"
  value       = module.logguardian.dashboard_url
}

output "manual_invocation_command" {
  description = "Command to manually trigger compliance check"
  value       = module.logguardian.manual_invocation_command
}
