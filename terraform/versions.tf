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
  default_tags {
    tags = {
      Product     = var.product_name
      Environment = var.environment
      Owner       = var.owner
      ManagedBy   = "Terraform"
      Module      = "LogGuardian"
    }
  }
}
