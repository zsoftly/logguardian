region      = "ca-central-1"
environment = "dev"

# Container image (defaults to current account's ECR if not specified)
# Uncomment to override:
# container_image = "123456789012.dkr.ecr.ca-central-1.amazonaws.com/logguardian:latest"

# Network Configuration
# Default: Uses the default VPC and public subnets
# To use custom VPC/subnets, uncomment and specify:
# vpc_id     = "vpc-xxxxxxxxx"
# subnet_ids = ["subnet-xxx", "subnet-yyy"]

assign_public_ip = true

cpu    = "256"
memory = "512"

enable_spot = true

log_retention_days = 30

config_rule_name = "cw-lg-retention-min"

enable_scheduling = true

# Schedule Configuration
# AWS Config evaluates every 24 hours minimum. Choose daily, weekly, or monthly.
# IMPORTANT: Stagger encryption and retention by 1 hour to avoid concurrent runs.

# OPTION 1: WEEKLY (Default - Recommended for cost optimization)
# Runs every Sunday - encryption at 3 AM UTC, retention at 4 AM UTC
encryption_schedule_expression = "cron(0 3 ? * SUN *)"
retention_schedule_expression  = "cron(0 4 ? * SUN *)"

# OPTION 2: DAILY (Fastest - Higher cost, runs every 24 hours)
# Runs every day - encryption at 2 AM UTC, retention at 3 AM UTC
# encryption_schedule_expression = "cron(0 2 * * ? *)"
# retention_schedule_expression  = "cron(0 3 * * ? *)"

# OPTION 3: MONTHLY (Lowest cost - Runs on 1st of each month)
# Runs monthly - encryption at 2 AM UTC, retention at 3 AM UTC
# encryption_schedule_expression = "cron(0 2 1 * ? *)"
# retention_schedule_expression  = "cron(0 3 1 * ? *)"

# OPTION 4: BUSINESS HOURS (Weekdays only during business hours)
# Runs Mon-Fri - encryption at 2 PM UTC (9 AM EST), retention at 3 PM UTC (10 AM EST)
# encryption_schedule_expression = "cron(0 14 ? * MON-FRI *)"
# retention_schedule_expression  = "cron(0 15 ? * MON-FRI *)"

# OPTION 5: CUSTOM STAGGERED (Different days)
# Encryption on Sundays, Retention on Mondays
# encryption_schedule_expression = "cron(0 2 ? * SUN *)"
# retention_schedule_expression  = "cron(0 2 ? * MON *)"

# Config rule names
encryption_config_rule = "logguardian-encryption-dev"
retention_config_rule  = "logguardian-retention-dev"

# Dry Run Mode
# When true: Only reports violations without making changes (recommended for testing)
# When false: Automatically remediates non-compliant log groups (production mode)
dry_run = true

tags = {
  Environment = "dev"
  Project     = "logguardian"
  ManagedBy   = "terraform"
  Component   = "ecs-fargate"
}
