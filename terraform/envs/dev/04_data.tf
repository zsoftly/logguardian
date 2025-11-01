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
}
