[{
  "name": "logguardian",
  "image": "${container_image}",
  "essential": true,
  "environment": [
    {
      "name": "AWS_REGION",
      "value": "${region}"
    },
    {
      "name": "AWS_DEFAULT_REGION",
      "value": "${region}"
    }
  ],
  "logConfiguration": {
    "logDriver": "awslogs",
    "options": {
      "awslogs-group": "${log_group_name}",
      "awslogs-region": "${region}",
      "awslogs-stream-prefix": "ecs"
    }
  }
}]
