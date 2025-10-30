# LogGuardian ECS/Fargate Deployment - Dev Environment

## Overview
ECS/Fargate infrastructure for running LogGuardian compliance automation in the dev account.

## Resources Deployed
- ECS Cluster: `logguardian-dev`
- Task Definition: `logguardian-dev`
- IAM Roles: Task + Execution
- Security Group: Egress-only for AWS API access
- CloudWatch Log Group: `/ecs/logguardian`
- Fargate Spot: 80/20 cost optimization

## Prerequisites
- AWS CLI configured
- Terraform >= 1.0
- ECR image: `769392325486.dkr.ecr.ca-central-1.amazonaws.com/logguardian:latest`
- AWS Config enabled in account

## Deployment

### Authenticate
\`\`\`bash
ztictl auth login zsoftly
# Select: zsoftly dev logguardian (769392325486)
# Select: AdministratorAccess

export AWS_PROFILE=zsoftly
export AWS_REGION=ca-central-1
\`\`\`

### Deploy
\`\`\`bash
cd terraform/environments/dev
terraform init
terraform plan
terraform apply
\`\`\`

### Test Execution
\`\`\`bash
# Get values from Terraform outputs
CLUSTER=$(terraform output -raw cluster_name)
TASK=$(terraform output -raw task_definition_family)
SG=$(terraform output -raw security_group_id)

# Run dry-run test
aws ecs run-task \
  --cluster $CLUSTER \
  --launch-type FARGATE \
  --task-definition $TASK \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-0cb3a166fffa03698,subnet-0026232dabc7d880d],securityGroups=[$SG],assignPublicIp=ENABLED}" \
  --overrides '{
    "containerOverrides":[{
      "name":"logguardian",
      "command":["--dry-run","--config-rule","cloudwatch-log-group-encrypted"],
      "environment":[{"name":"AWS_REGION","value":"ca-central-1"}]
    }]
  }'

# Check logs
aws logs tail /ecs/logguardian --since 5m
\`\`\`

## Cost
- Fargate Spot: ~$15-20/month
- CloudWatch Logs: ~$1-2/month
- ECR Storage: <$1/month
- **Total: ~$20/month**

## Network Architecture
Uses **public subnets** with direct internet gateway access (no NAT Gateway).
- ✅ Cost savings: -$32/month
- ✅ Suitable for dev/test
- ⚠️ Consider private subnets + NAT for production

## Outputs
- `cluster_name`: ECS cluster name
- `task_definition_arn`: Full task definition ARN
- `task_definition_family`: Task family name
- `task_role_arn`: IAM role for task permissions
- `execution_role_arn`: IAM role for ECS agent
- `security_group_id`: Security group ID
- `log_group_name`: CloudWatch log group

## Troubleshooting

### Image Pull Errors
If tasks fail with "CannotPullContainerError":
\`\`\`bash
# Rebuild and push image
cd ~/logguardian
docker build -t logguardian:latest .
docker tag logguardian:latest 769392325486.dkr.ecr.ca-central-1.amazonaws.com/logguardian:latest
aws ecr get-login-password --region ca-central-1 | docker login --username AWS --password-stdin 769392325486.dkr.ecr.ca-central-1.amazonaws.com
docker push 769392325486.dkr.ecr.ca-central-1.amazonaws.com/logguardian:latest
\`\`\`

### No Logs
Check execution role has `AmazonECSTaskExecutionRolePolicy` attached.

### Config Rule Not Found
Create AWS Config rule first:
\`\`\`bash
aws configservice put-config-rule --config-rule '{
  "ConfigRuleName": "cloudwatch-log-group-encrypted",
  "Source": {"Owner": "AWS", "SourceIdentifier": "CLOUDWATCH_LOG_GROUP_ENCRYPTED"},
  "Scope": {"ComplianceResourceTypes": ["AWS::Logs::LogGroup"]}
}'
\`\`\`
