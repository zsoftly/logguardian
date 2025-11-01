# LogGuardian Terraform Infrastructure

## Overview

This directory contains the Terraform code to deploy the LogGuardian application on AWS using ECS Fargate. It is designed to be modular and reusable across different environments.

The infrastructure creates the necessary AWS resources for running LogGuardian as a scheduled, event-driven task, optimizing for cost and security.

## Core Architecture

The Terraform configuration deploys the following core components:

-   **ECS Fargate Cluster:** A serverless compute engine for containers.
-   **ECS Task Definition:** Configured to run the LogGuardian container image using Fargate Spot for cost savings.
-   **IAM Roles:** Includes a task role, execution role, and EventBridge role, all scoped with least-privilege permissions.
-   **Security Group:** A default egress-only security group to allow the container to make AWS API calls.
-   **CloudWatch Log Group:** Centralized logging for the container with configurable retention.
-   **EventBridge Scheduled Rules:** Triggers the ECS task based on a defined cron schedule.

### Network Design

By default, the infrastructure is configured to use an account's **default VPC and its public subnets**. This approach simplifies deployment and avoids the need for a NAT Gateway, making it cost-effective and suitable for environments where a custom VPC is not required. For production or environments with specific networking requirements, the VPC and subnet IDs can be customized.

## Example Configuration: `dev` Environment

Environments are configured in their respective subdirectories (e.g., `envs/dev`). The following is an example from the `dev` environment's `env.auto.tfvars` file, which defines the specific settings for that deployment.

### Required Variables

These variables must be set for each environment.

```hcl
# terraform/envs/dev/env.auto.tfvars

region          = "ca-central-1"
environment     = "dev"
container_image = "ACCOUNT_ID.dkr.ecr.ca-central-1.amazonaws.com/logguardian:latest"
```

**Note:** The `ACCOUNT_ID` in `container_image` must be replaced with the target AWS account number. It can be retrieved with the AWS CLI:
`aws sts get-caller-identity --query Account --output text`

### Custom Network (Optional)

To use a specific VPC instead of the default, you can provide the VPC and subnet IDs.

```hcl
# terraform/envs/dev/env.auto.tfvars

# vpc_id     = "vpc-xxxxxxxxx"
# subnet_ids = ["subnet-xxx", "subnet-yyy"]
```

## Event-Driven Scheduling

LogGuardian is designed to run periodically, not continuously. It performs compliance checks that align with the 24-hour evaluation cycle of AWS Config. The schedule is managed by EventBridge and can be customized based on cost and compliance needs.

For the `dev` environment, we use a **weekly** schedule to balance cost and regular testing. The encryption and retention checks are staggered by one hour to prevent concurrent runs.

-   **Encryption Check:** `cron(0 3 ? * SUN *)` - Sunday at 3 AM UTC
-   **Retention Check:** `cron(0 4 ? * SUN *)` - Sunday at 4 AM UTC

Other scheduling strategies can be configured in the environment's `.tfvars` file. See `05_variables.tf` for more details on available options like `DAILY`, `MONTHLY`, or `BUSINESS_HOURS`. To disable scheduling entirely, set `enable_scheduling = false`.

## Deployment & Testing

The following steps outline the general process for deploying the infrastructure, using the `dev` environment as an example.

### 1. Authenticate and Initialize

First, ensure your AWS CLI is authenticated to the target account. Then, navigate to the environment directory and initialize Terraform.

```bash
# Authenticate to the desired AWS account
# Example: ztictl auth login zsoftly

cd terraform/envs/dev
terraform init
```

### 2. Plan and Apply

Review the planned changes and apply them to create the resources.

```bash
terraform plan
terraform apply
```

### 3. Manual Test Execution

After deployment, you can manually trigger the ECS task to verify its operation. The required network and task details can be retrieved from Terraform outputs.

```bash
# Get required values from Terraform outputs
CLUSTER=$(terraform output -raw cluster_name)
TASK=$(terraform output -raw task_definition_family)
SG=$(terraform output -raw security_group_id)
SUBNETS=$(terraform output -json subnet_ids | jq -r 'join(",")')
LOG_GROUP=$(terraform output -raw log_group_name)

# Run a test task with default settings
aws ecs run-task \
  --cluster "$CLUSTER" \
  --launch-type FARGATE \
  --task-definition "$TASK" \
  --network-configuration "awsvpcConfiguration={subnets=[$SUBNETS],securityGroups=[$SG],assignPublicIp=ENABLED}"

# Follow the logs in CloudWatch
aws logs tail "$LOG_GROUP" --since 5m --follow
```

## Troubleshooting

Common issues and their resolutions are documented below.

### CannotPullContainerError

This error indicates the ECS agent cannot pull the container image from ECR.
1.  **Verify Image URI:** Ensure the `container_image` variable in your `.tfvars` file is correct and the image tag exists in ECR.
2.  **Check ECR Permissions:** The ECS Task Execution Role must have permissions to pull from ECR (`AmazonECSTaskExecutionRolePolicy`).
3.  **Authentication:** If pushing a new image, ensure you are authenticated to the ECR registry.

### Task Fails with No Logs

If the task fails immediately and no logs appear in CloudWatch, the ECS Execution Role likely lacks permissions to write logs.
-   Verify the role has the `AmazonECSTaskExecutionRolePolicy` attached.
-   Check the trust relationship for the role to ensure `ecs-tasks.amazonaws.com` is a trusted principal.

### Config Rule Not Found

LogGuardian depends on AWS Config rules being present in the account. If a task fails because a rule is missing, create it using the AWS CLI or Console. Example:
```bash
aws configservice put-config-rule --config-rule '{
  "ConfigRuleName": "cw-lg-retention-min",
  "Source": {"Owner": "CUSTOM_LAMBDA"},
  "Scope": {"ComplianceResourceTypes": ["AWS::Logs::LogGroup"]}
}'
```

## File Structure

Each environment follows a consistent numbered file structure for clarity:

```
terraform/envs/{environment}/
├── 01_backend.tf        # S3 backend configuration
├── 02_provider.tf       # Provider and required versions
├── 03_locals.tf         # Computed local values
├── 04_data.tf           # Data sources (VPC, subnets)
├── 05_variables.tf      # Variable definitions
├── 06_main.tf           # ECS resources
├── 07_iam.tf            # IAM roles and policies
├── 08_outputs.tf        # Output values
├── 09_eventbridge.tf    # EventBridge scheduled rules
└── env.auto.tfvars      # Environment-specific values
```

## Cost Estimate (dev environment)

- Fargate Spot: ~$15-20/month
- CloudWatch Logs: ~$1-2/month
- ECR Storage: <$1/month
- **Total: ~$20/month**

Using the default VPC and public subnets saves ~$32/month by avoiding NAT Gateway costs.
