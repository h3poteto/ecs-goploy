# ecs-goploy
[![CircleCI](https://circleci.com/gh/h3poteto/ecs-goploy.svg?style=svg)](https://circleci.com/gh/h3poteto/ecs-goploy)
[![GitHub release](http://img.shields.io/github/release/h3poteto/ecs-goploy.svg?style=flat-square)](https://github.com/h3poteto/ecs-goploy/releases)
[![GoDoc](https://godoc.org/github.com/h3poteto/ecs-goploy/deploy?status.svg)](https://godoc.org/github.com/h3poteto/ecs-goploy/deploy)

`ecs-goploy` is a re-implementation of [ecs-deploy](https://github.com/silinternational/ecs-deploy) in Golang.


This is a command line tool, but you can use `deploy` as a package.
So when you write own deploy script for AWS ECS, you can embed `deploy` package in your golang source code and customize deploy recipe.
Please check [godoc](https://godoc.org/github.com/h3poteto/ecs-goploy/deploy).


# Install

Get binary from github:

```
$ wget https://github.com/h3poteto/ecs-goploy/releases/download/v0.3.4/ecs-goploy_v0.3.4_linux_amd64.zip
$ unzip ecs-goploy_v0.3.4_linux_amd64.zip
$ ./ecs-goploy --help
```

# Usage
## Deploy an ECS Service


```
$ ./ecs-goploy update service --help
Deploy an ECS Service

Usage:
  ecs-goploy update service [flags]

Flags:
  -c, --cluster string           Name of ECS cluster
      --enable-rollback          Rollback task definition if new version is not running before TIMEOUT
  -h, --help                     help for service
  -i, --image string             Name of Docker image to run, ex: repo/image:latest
  -p, --profile string           AWS Profile to use
  -r, --region string            AWS Region Name
  -n, --service-name string      Name of service to deploy
      --skip-check-deployments   Skip checking deployments when detect whether deploy completed
  -d, --task-definition string   Name of base task definition to deploy. Family and revision (family:revision) or full ARN
  -t, --timeout int              Timeout seconds. Script monitors ECS Service for new task definition to be running (default 300)
```

## Run Task

```
$ ./ecs-goploy run task --help
Run task on ECS

Usage:
  ecs-goploy run task [flags]

Flags:
  -c, --cluster string           Name of ECS cluster
      --command string           Task command which run on ECS
  -n, --container-name string    Name of the container for override task definition
  -h, --help                     help for task
  -p, --profile string           AWS Profile to use
  -r, --region string            AWS Region Name
  -d, --task-definition string   Name of task definition to run task. Family and revision (family:revision) or full ARN
  -t, --timeout int              Timeout seconds (default 300)
```

## Update Scheduled Task

```
$ ./ecs-goploy update scheduled-task --help
Update ECS Scheduled Task

Usage:
  ecs-goploy update scheduled-task [flags]

Flags:
  -c, --count int                Count of the task (default 1)
  -h, --help                     help for scheduled-task
  -n, --name string              Name of scheduled task
  -d, --task-definition string   Name of task definition to update scheduled task. Family and revision (family:revision) or full ARN
```


# Configuration
## AWS Configuration

`ecs-goploy` calls AWS API via aws-skd-go, so you need export environment variables:

```
$ export AWS_ACCESS_KEY_ID=XXXXX
$ export AWS_SECRET_ACCESS_KEY=XXXXX
$ export AWS_DEFAULT_REGION=XXXXX
```

or set your credentials in `$HOME/.aws/credentials`:

```
[default]
aws_access_key_id = XXXXX
aws_secret_access_key = XXXXX
```

or prepare IAM Role or IAM Task Role.

AWS region can be set command argument: `--region`.

## AWS IAM Policy

Below is a basic IAM Policy required for ecs-goploy.

```
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowUserToECSDeploy",
      "Effect": "Allow",
      "Action": [
        "ecr:DescribeRepositories",
        "ecr:DescribeImages",
        "ecs:DescribeServices",
        "ecs:DescribeTaskDefinition",
        "ecs:RegisterTaskDefinition",
        "ecs:UpdateService",
        "ecs:RunTask",
        "ecs:DescribeTasks",
        "ecs:ListTasks",
        "iam:PassRole"
      ],
      "Resource": "*"
    }
  ]
}
```

# License

The package is available as open source under the terms of the [MIT License](https://opensource.org/licenses/MIT).
