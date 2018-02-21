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

```
$ ./ecs-goploy help
Deploy commands for ECS

Usage:
  ecs-goploy [command]

Available Commands:
  help        Help about any command
  service     Service deploy to ECS
  task        Run task on ECS
  version     Print the version number

Flags:
  -h, --help   help for ecs-goploy

Use "ecs-goploy [command] --help" for more information about a command.

$ ./ecs-goploy service --help
Service deploy to ECS

Usage:
  ecs-goploy service [flags]

Flags:
  -c, --cluster string           Name of ECS cluster
      --enable-rollback          Rollback task definition if new version is not running before TIMEOUT
  -h, --help                     help for service
  -i, --image string             Name of Docker image to run, ex: repo/image:latest
  -p, --profile string           AWS Profile to use
  -r, --region string            AWS Region Name
  -n, --service-name string      Name of service to deploy
  -d, --task-definition string   Name of base task definition to deploy. Family and revision (family:revision) or full ARN
  -t, --timeout int              Timeout seconds. Script monitors ECS Service for new task definition to be running (default 300)

$ ./ecs-goploy task --help                                                                                                                                                              [master]
Run task on ECS

Usage:
  ecs-goploy task [flags]

Flags:
  -c, --cluster string           Name of ECS cluster
      --command string           Task command which run on ECS
  -n, --container-name string    Name of the container for override task definition
  -h, --help                     help for task
  -i, --image string             Name of Doker image to run, ex: repo/image:latest
  -p, --profile string           AWS Profile to use
  -r, --region string            AWS Region Name
  -d, --task-definition string   Name of base task definition to run task. Family and revision (family:revision) or full ARN
  -t, --timeout int              Timeout seconds (default 300)
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
        "ecs:DescribeTasks"
      ],
      "Resource": "*"
    }
  ]
}
```

# License

The package is available as open source under the terms of the [MIT License](https://opensource.org/licenses/MIT).
