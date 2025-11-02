# Docker Deployment Guide

## Prerequisites

- Docker 20.10 or later
- AWS credentials with appropriate IAM permissions

## Installation

Pull the image directly from GitHub Container Registry:

```bash
docker pull ghcr.io/zsoftly/logguardian:1.4.1
docker pull ghcr.io/zsoftly/logguardian:latest
```

No authentication required - the image is publicly accessible.

### Image Tags

| Tag Format | Description | Example |
|------------|-------------|---------|
| `X.Y.Z` | Specific version | `1.4.1` |
| `X.Y` | Latest patch version | `1.4` |
| `X` | Latest minor version | `1` |
| `latest` | Latest stable release | `latest` |

**Note**: Docker tags use semantic versioning without the `v` prefix.

## Configuration

### Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `CONFIG_RULE_NAME` | AWS Config rule name | Yes | - |
| `AWS_REGION` | AWS region | Yes | - |
| `BATCH_SIZE` | Resources per batch | No | `10` |
| `DRY_RUN` | Preview mode | No | `false` |

### Command-Line Options

```
--config-rule <name>    AWS Config rule name
--region <region>       AWS region
--batch-size <n>        Batch size (1-100)
--dry-run              Enable preview mode
--profile <name>        AWS profile name
--assume-role <arn>     IAM role ARN to assume
--output <format>       Output format (json|text)
--verbose              Enable debug logging
```

## Usage

### Local Execution

```bash
# Preview mode
docker run --rm \
  -v ~/.aws:/home/logguardian/.aws:ro \
  -e AWS_PROFILE=default \
  ghcr.io/zsoftly/logguardian:latest \
  --config-rule cw-lg-retention-min \
  --region ca-central-1 \
  --dry-run

# Production mode
docker run --rm \
  -e AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
  -e AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
  -e AWS_REGION=ca-central-1 \
  ghcr.io/zsoftly/logguardian:latest \
  --config-rule cw-lg-retention-min \
  --batch-size 20
```

## AWS ECS Deployment

### Task Definition

```json
{
  "family": "logguardian-retention",
  "taskRoleArn": "arn:aws:iam::ACCOUNT_ID:role/logguardian-task-role",
  "executionRoleArn": "arn:aws:iam::ACCOUNT_ID:role/ecsTaskExecutionRole",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "containerDefinitions": [{
    "name": "logguardian",
    "image": "ghcr.io/zsoftly/logguardian:1.4.1",
    "essential": true,
    "environment": [
      {"name": "CONFIG_RULE_NAME", "value": "cw-lg-retention-min"},
      {"name": "AWS_REGION", "value": "ca-central-1"},
      {"name": "BATCH_SIZE", "value": "20"}
    ],
    "logConfiguration": {
      "logDriver": "awslogs",
      "options": {
        "awslogs-group": "/ecs/logguardian",
        "awslogs-region": "ca-central-1",
        "awslogs-stream-prefix": "retention"
      }
    }
  }]
}
```

### Register Task

```bash
aws ecs register-task-definition \
  --cli-input-json file://logguardian-task.json \
  --region ca-central-1
```

### Schedule with EventBridge

```bash
# Create schedule rule
aws events put-rule \
  --name logguardian-retention-daily \
  --schedule-expression "cron(0 2 * * ? *)" \
  --region ca-central-1

# Add ECS target
aws events put-targets \
  --rule logguardian-retention-daily \
  --targets '{
    "Id": "1",
    "Arn": "arn:aws:ecs:ca-central-1:ACCOUNT_ID:cluster/logguardian-cluster",
    "RoleArn": "arn:aws:iam::ACCOUNT_ID:role/ecsEventsRole",
    "EcsParameters": {
      "TaskDefinitionArn": "arn:aws:ecs:ca-central-1:ACCOUNT_ID:task-definition/logguardian-retention",
      "LaunchType": "FARGATE",
      "NetworkConfiguration": {
        "awsvpcConfiguration": {
          "Subnets": ["subnet-xxx"],
          "SecurityGroups": ["sg-xxx"],
          "AssignPublicIp": "ENABLED"
        }
      }
    }
  }' \
  --region ca-central-1
```

### Run On-Demand Task

```bash
aws ecs run-task \
  --cluster logguardian-cluster \
  --task-definition logguardian-retention \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-xxx],securityGroups=[sg-xxx],assignPublicIp=ENABLED}" \
  --region ca-central-1
```

## IAM Permissions

### Task Execution Role

Required for ECS to pull images and write logs:

```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {"Service": "ecs-tasks.amazonaws.com"},
    "Action": "sts:AssumeRole"
  }]
}
```

Attach AWS managed policies:
- `arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy`

### Task Role

Required for LogGuardian to access AWS services:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "config:GetComplianceDetailsByConfigRule",
        "config:PutEvaluations"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:DescribeLogGroups",
        "logs:PutRetentionPolicy",
        "logs:AssociateKmsKey"
      ],
      "Resource": "arn:aws:logs:*:*:log-group:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "kms:DescribeKey",
        "kms:CreateGrant",
        "kms:Decrypt"
      ],
      "Resource": "arn:aws:kms:*:*:key/*"
    }
  ]
}
```

## Troubleshooting

### Image Pull Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `manifest unknown` | Invalid tag format | Verify tag format (no `v` prefix) - use `1.4.1` not `v1.4.1` |
| `not found` | Incorrect image name or tag | Verify image path: `ghcr.io/zsoftly/logguardian:latest` |

### Runtime Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `region is required` | Missing region configuration | Set `AWS_REGION` or use `--region` flag |
| `config rule name is required` | Missing rule configuration | Set `CONFIG_RULE_NAME` or use `--config-rule` flag |
| `NoCredentialsError` | Missing AWS credentials | Mount credentials, set environment variables, or use ECS task role |

## Building from Source

```bash
git clone https://github.com/zsoftly/logguardian.git
cd logguardian
docker build -t logguardian:local .
docker run --rm logguardian:local --help
```

## Support

- **Documentation**: https://github.com/zsoftly/logguardian
- **Issues**: https://github.com/zsoftly/logguardian/issues
- **Releases**: https://github.com/zsoftly/logguardian/releases
