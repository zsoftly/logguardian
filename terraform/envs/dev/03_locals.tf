locals {
  account_id = data.aws_caller_identity.current.account_id
  vpc_id     = data.aws_vpc.selected.id
  subnet_ids = var.subnet_ids != null ? var.subnet_ids : data.aws_subnets.selected.ids

  name_prefix = "logguardian-${var.environment}"

  log_group_name = "/ecs/logguardian-${var.environment}"

  # Default container image uses current account's ECR
  default_container_image = "${local.account_id}.dkr.ecr.${var.region}.amazonaws.com/logguardian:latest"
  container_image         = var.container_image != null ? var.container_image : local.default_container_image
}
