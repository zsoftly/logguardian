# Docker Usage Guide

## Quick Start

### Pull the Image

```bash
# Latest version
docker pull ghcr.io/zsoftly/logguardian:latest

# Specific version
docker pull ghcr.io/zsoftly/logguardian:v1.4.0
```

### Run Locally

```bash
# Help
docker run --rm ghcr.io/zsoftly/logguardian:latest --help

# Dry-run mode (preview changes)
docker run --rm \
  -e AWS_REGION=ca-central-1 \
  -e CONFIG_RULE_NAME=your-config-rule \
  -e DRY_RUN=true \
  ghcr.io/zsoftly/logguardian:latest

# With AWS credentials
docker run --rm \
  -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
  -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
  -e AWS_REGION=ca-central-1 \
  ghcr.io/zsoftly/logguardian:latest \
  --config-rule my-config-rule \
  --batch-size 20
```

## Deploy to AWS ECS

LogGuardian requires two AWS Config rules to be evaluated:
1. **Retention Compliance** - Ensures log groups have proper retention periods
2. **Encryption Compliance** - Ensures log groups are encrypted with KMS

You'll need to create two ECS task definitions, one for each rule.

### 1. Create Task Definitions

#### Retention Task Definition
Create `logguardian-retention-task.json`:

```json
{
  "family": "logguardian-retention",
  "taskRoleArn": "arn:aws:iam::ACCOUNT_ID:role/logguardian-task-role",
  "executionRoleArn": "arn:aws:iam::ACCOUNT_ID:role/ecsTaskExecutionRole",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "containerDefinitions": [
    {
      "name": "logguardian",
      "image": "ghcr.io/zsoftly/logguardian:latest",
      "essential": true,
      "environment": [
        {
          "name": "CONFIG_RULE_NAME",
          "value": "logguardian-log-retention"
        },
        {
          "name": "AWS_REGION",
          "value": "ca-central-1"
        },
        {
          "name": "BATCH_SIZE",
          "value": "20"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/logguardian",
          "awslogs-region": "ca-central-1",
          "awslogs-stream-prefix": "retention"
        }
      }
    }
  ]
}
```

#### Encryption Task Definition
Create `logguardian-encryption-task.json`:

```json
{
  "family": "logguardian-encryption",
  "taskRoleArn": "arn:aws:iam::ACCOUNT_ID:role/logguardian-task-role",
  "executionRoleArn": "arn:aws:iam::ACCOUNT_ID:role/ecsTaskExecutionRole",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "containerDefinitions": [
    {
      "name": "logguardian",
      "image": "ghcr.io/zsoftly/logguardian:latest",
      "essential": true,
      "environment": [
        {
          "name": "CONFIG_RULE_NAME",
          "value": "logguardian-log-encryption"
        },
        {
          "name": "AWS_REGION",
          "value": "ca-central-1"
        },
        {
          "name": "BATCH_SIZE",
          "value": "20"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/logguardian",
          "awslogs-region": "ca-central-1",
          "awslogs-stream-prefix": "encryption"
        }
      }
    }
  ]
}
```

### 2. Register Task Definitions

```bash
# Register retention task
aws ecs register-task-definition \
  --cli-input-json file://logguardian-retention-task.json \
  --region ca-central-1

# Register encryption task
aws ecs register-task-definition \
  --cli-input-json file://logguardian-encryption-task.json \
  --region ca-central-1
```

### 3. Create ECS Cluster (if needed)

```bash
aws ecs create-cluster \
  --cluster-name logguardian-cluster \
  --region ca-central-1
```

### 4. Run as Scheduled Tasks (EventBridge)

```bash
# Create EventBridge rule for retention check (runs daily at 2 AM UTC)
aws events put-rule \
  --name logguardian-retention-daily \
  --schedule-expression "cron(0 2 * * ? *)" \
  --region ca-central-1

# Create EventBridge rule for encryption check (runs daily at 3 AM UTC)
aws events put-rule \
  --name logguardian-encryption-daily \
  --schedule-expression "cron(0 3 * * ? *)" \
  --region ca-central-1

# Create ECS target for retention rule
aws events put-targets \
  --rule logguardian-retention-daily \
  --targets "Id"="1","Arn"="arn:aws:ecs:ca-central-1:ACCOUNT_ID:cluster/logguardian-cluster","RoleArn"="arn:aws:iam::ACCOUNT_ID:role/ecsEventsRole","EcsParameters"="{\"TaskDefinitionArn\":\"arn:aws:ecs:ca-central-1:ACCOUNT_ID:task-definition/logguardian-retention\",\"LaunchType\":\"FARGATE\",\"NetworkConfiguration\":{\"awsvpcConfiguration\":{\"Subnets\":[\"subnet-xxx\"],\"SecurityGroups\":[\"sg-xxx\"],\"AssignPublicIp\":\"ENABLED\"}}}" \
  --region ca-central-1

# Create ECS target for encryption rule
aws events put-targets \
  --rule logguardian-encryption-daily \
  --targets "Id"="1","Arn"="arn:aws:ecs:ca-central-1:ACCOUNT_ID:cluster/logguardian-cluster","RoleArn"="arn:aws:iam::ACCOUNT_ID:role/ecsEventsRole","EcsParameters"="{\"TaskDefinitionArn\":\"arn:aws:ecs:ca-central-1:ACCOUNT_ID:task-definition/logguardian-encryption\",\"LaunchType\":\"FARGATE\",\"NetworkConfiguration\":{\"awsvpcConfiguration\":{\"Subnets\":[\"subnet-xxx\"],\"SecurityGroups\":[\"sg-xxx\"],\"AssignPublicIp\":\"ENABLED\"}}}" \
  --region ca-central-1
```

### 5. Run One-Time Tasks

```bash
# Run retention check
aws ecs run-task \
  --cluster logguardian-cluster \
  --task-definition logguardian-retention \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-xxx],securityGroups=[sg-xxx],assignPublicIp=ENABLED}" \
  --region ca-central-1

# Run encryption check
aws ecs run-task \
  --cluster logguardian-cluster \
  --task-definition logguardian-encryption \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-xxx],securityGroups=[sg-xxx],assignPublicIp=ENABLED}" \
  --region ca-central-1
```

### 6. Create IAM Task Role

The task needs the same permissions as the Lambda function. Create `logguardian-task-role.json`:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

Create the role and attach policies:

```bash
# Create role
aws iam create-role \
  --role-name logguardian-task-role \
  --assume-role-policy-document file://logguardian-task-role.json

# Attach policies (same as Lambda function)
aws iam attach-role-policy \
  --role-name logguardian-task-role \
  --policy-arn arn:aws:iam::aws:policy/service-role/AWSConfigRole

aws iam attach-role-policy \
  --role-name logguardian-task-role \
  --policy-arn arn:aws:iam::aws:policy/CloudWatchLogsFullAccess

# Add KMS permissions if using encryption
aws iam put-role-policy \
  --role-name logguardian-task-role \
  --policy-name KMSAccess \
  --policy-document '{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Action": [
          "kms:CreateGrant",
          "kms:Decrypt",
          "kms:DescribeKey"
        ],
        "Resource": "*"
      }
    ]
  }'
```

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `AWS_REGION` | AWS region | Yes |
| `CONFIG_RULE_NAME` | Config rule to evaluate | Yes |
| `BATCH_SIZE` | Resources per batch (1-100) | No (default: 10) |
| `DRY_RUN` | Preview mode (true/false) | No (default: false) |

## Command Line Options

```bash
--config-rule <name>    Config rule name
--region <region>       AWS region  
--batch-size <1-100>    Batch size
--dry-run              Preview mode
--output <json|text>    Output format
--verbose              Debug logging
```

## Build Locally

```bash
# Clone and build
git clone https://github.com/zsoftly/logguardian.git
cd logguardian
docker build -t logguardian:local .
```