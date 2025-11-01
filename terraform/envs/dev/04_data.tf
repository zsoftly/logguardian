data "aws_caller_identity" "current" {}

data "aws_vpc" "selected" {
  default = var.vpc_id == null ? true : null
  id      = var.vpc_id
}

data "aws_subnets" "selected" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.selected.id]
  }

  filter {
    name   = "map-public-ip-on-launch"
    values = ["true"]
  }

  lifecycle {
    postcondition {
      condition     = length(self.ids) > 0
      error_message = "No public subnets found in VPC ${data.aws_vpc.selected.id}. Please check the VPC configuration or provide subnet IDs manually via the subnet_ids variable."
    }
  }
}
