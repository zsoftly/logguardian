locals {
  account_id = data.aws_caller_identity.current.account_id
  vpc_id     = data.aws_vpc.selected.id
  subnet_ids = var.subnet_ids != null ? var.subnet_ids : data.aws_subnets.selected.ids

  name_prefix = "logguardian-${var.environment}"

  log_group_name = "/ecs/logguardian-${var.environment}"
}
