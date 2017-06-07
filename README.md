# ecs-goploy
[![CircleCI](https://circleci.com/gh/crowdworks/ecs-goploy.svg?style=svg)](https://circleci.com/gh/crowdworks/ecs-goploy)

`ecs-goploy` is a re-implementation of [ecs-deploy](https://github.com/silinternational/ecs-deploy) in Golang.

__NOTICE: This package is ACTIVELY under development. API (both public and internal) may change suddenly.__

# Install

Get binary from github.

# Usage

```
$ ./ecs-goploy help
Deploy commands for ECS

Usage:
  ecs-goploy [command]

Available Commands:
  deploy      Deploy ECS
  help        Help about any command
  version     Print the version number

Flags:
  -h, --help   help for ecs-goploy

Use "ecs-goploy [command] --help" for more information about a command.

$ ./ecs-goploy deploy --help
Deploy ECS

Usage:
  ecs-goploy deploy [flags]

Flags:
  -c, --cluster string        Name of ECS cluster
      --enable-rollback       Rollback task definition if new version is not running before TIMEOUT
  -h, --help                  help for deploy
  -i, --image string          Name of Docker image to run, ex: repo/image:latest
  -p, --profile string        AWS Profile to use
  -r, --region string         AWS Region Name
  -n, --service-name string   Name of service to deploy
  -t, --timeout int           Timeout seconds. Script monitors ECS Service for new task definition to be running (default 300)
```

# Configuration

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


# TODO
- [ ] Tests
- [ ] Deploy when a task definition is provided in command argument: `--task-definition`
