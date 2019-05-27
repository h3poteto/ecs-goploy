# ecs-goploy
[![Build Status](https://travis-ci.com/h3poteto/ecs-goploy.svg?branch=master)](https://travis-ci.com/h3poteto/ecs-goploy)
[![GitHub release](http://img.shields.io/github/release/h3poteto/ecs-goploy.svg?style=flat-square)](https://github.com/h3poteto/ecs-goploy/releases)
[![GoDoc](https://godoc.org/github.com/h3poteto/ecs-goploy/deploy?status.svg)](https://godoc.org/github.com/h3poteto/ecs-goploy/deploy)

`ecs-goploy` is a re-implementation of [ecs-deploy](https://github.com/silinternational/ecs-deploy) in Golang.


This is a command line tool, but you can use `deploy` as a package.
So when you write own deploy script for AWS ECS, you can embed `deploy` package in your golang source code and customize deploy recipe.
Please check [godoc](https://godoc.org/github.com/h3poteto/ecs-goploy/deploy).


# Install

Get binary from github:

```
$ wget https://github.com/h3poteto/ecs-goploy/releases/download/v0.5.0/ecs-goploy_v0.5.0_linux_amd64.zip
$ unzip ecs-goploy_v0.5.0_linux_amd64.zip
$ ./ecs-goploy --help
```

# Usage

```
$ ./ecs-goploy --help
Deploy commands for ECS

Usage:
  ecs-goploy [command]

Available Commands:
  help        Help about any command
  run         Run command
  update      Update some ECS resource
  version     Print the version number

Flags:
  -h, --help             help for ecs-goploy
      --profile string   AWS profile (detault is none, and use environment variables)
      --region string    AWS region (default is none, and use AWS_DEFAULT_REGION)
  -v, --verbose          Enable verbose mode

Use "ecs-goploy [command] --help" for more information about a command.
```

## Deploy an ECS Service

Please specify cluser, service name and image(family:revision).

```
$ ./ecs-goploy update service --cluster my-cluster --service-name my-service --image nginx:stable --skip-check-deployments --enable-rollback
```

If you specify `--base-task-definition`, ecs-goploy updates the task definition with the image and deploy ecs service.
If you does not specify `--base-task-definition`, ecs-goploy get current task definition of the service, and update with the image, and deploy ecs service.

## Run Task

At first, you must update the task definition which is used to run ecs task.
After that, you can run ecs task.

```
$ NEW_TASK_DEFINITION=`./ecs-goploy update task-definition --base-task-definition my-task-definition:1 --image nginx:stable`
$ ./ecs-goploy run task --cluster my-cluster --container-name web --task-definition $NEW_TASK_DEFINITION --command "some commands"
```

## Update Scheduled Task

At first, you must update the task definition which is used to run scheduled task.
After that, you can update the scheduled task.

```
$ NEW_TASK_DEFINITION=`./ecs-goploy update task-definition --base-task-definition my-task-definition:1 --image nginx:stable`
$ ./ecs-goploy update scheduled-task --count 1 --name schedule-name --task-definition $NEW_TASK_DEFINITION
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
        "events:DescribeRule",
        "events:ListTargetsByRule",
        "events:PutTargets",
        "iam:PassRole"
      ],
      "Resource": "*"
    }
  ]
}
```

# License

The package is available as open source under the terms of the [MIT License](https://opensource.org/licenses/MIT).
