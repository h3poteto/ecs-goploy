package deploy

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	shellwords "github.com/mattn/go-shellwords"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Task has target ECS information, client of aws-sdk-go, command and timeout seconds.
type Task struct {
	awsECS ecsiface.ECSAPI

	// Name of ECS cluster.
	Cluster string

	// Name of the container for override task definition.
	Name string

	// Name of base task definition for run task.
	BaseTaskDefinition string

	// TaskDefinition struct to call aws API.
	TaskDefinition *TaskDefinition

	// Task command which run on ECS.
	Command []*string

	// Wait time when run task.
	// This script monitors ECS task for new task definition to be running after call run task API.
	Timeout time.Duration
	// EC2 or Fargate
	LaunchType string
	// If you set Fargate as launch type, you have to set your subnet IDs.
	// Because Fargate demands awsvpc as network configuration, so subnet IDs are required.
	Subnets []*string
	// If you want to attach the security groups to ENI of the task, please set this.
	SecurityGroups []*string
	// If you don't enable this flag, the task access the internet throguth NAT gateway.
	// Please read more information: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-networking.html
	AssignPublicIP string
	verbose        bool
}

// NewTask returns a new Task struct, and initialize aws ecs API client.
// If you want to run the task as Fargate, please provide fargate flag to true, and your subnet IDs for awsvpc.
// If you don't want to run the task as Fargate, please provide empty string for subnetIDs.
func NewTask(cluster, name, command, baseTaskDefinition string, fargate bool, subnetIDs, securityGroupIDs string, timeout time.Duration, profile, region string, verbose bool) (*Task, error) {
	if baseTaskDefinition == "" {
		return nil, errors.New("task definition is required")
	}
	awsECS := ecs.New(session.New(), newConfig(profile, region))
	taskDefinition := NewTaskDefinition(profile, region, verbose)
	if !verbose {
		log.SetLevel(log.ErrorLevel)
	}
	p := shellwords.NewParser()
	commands, err := p.Parse(command)
	if err != nil {
		return nil, errors.Wrap(err, "Parse error in a task command")
	}
	var cmd []*string
	for _, c := range commands {
		cmd = append(cmd, aws.String(c))
	}
	launchType := "EC2"
	assignPublicIP := "DISABLED"
	if fargate {
		launchType = "FARGATE"
		assignPublicIP = "ENABLED"
	}
	subnets := []*string{}
	for _, s := range strings.Split(subnetIDs, ",") {
		if len(s) > 0 {
			subnets = append(subnets, aws.String(s))
		}
	}
	securityGroups := []*string{}
	for _, g := range strings.Split(securityGroupIDs, ",") {
		if len(g) > 0 {
			securityGroups = append(securityGroups, aws.String(g))
		}
	}

	return &Task{
		awsECS:             awsECS,
		Cluster:            cluster,
		Name:               name,
		BaseTaskDefinition: baseTaskDefinition,
		TaskDefinition:     taskDefinition,
		Command:            cmd,
		Timeout:            timeout,
		LaunchType:         launchType,
		Subnets:            subnets,
		SecurityGroups:     securityGroups,
		AssignPublicIP:     assignPublicIP,
		verbose:            verbose,
	}, nil
}

// RunTask calls run-task API.
func (t *Task) RunTask(taskDefinition *ecs.TaskDefinition) ([]*ecs.Task, error) {
	ctx, cancel := context.WithCancel(context.Background())
	if t.Timeout != 0 {
		ctx, cancel = context.WithTimeout(context.Background(), t.Timeout)
	}
	defer cancel()

	containerOverride := &ecs.ContainerOverride{
		Command: t.Command,
		Name:    aws.String(t.Name),
	}

	override := &ecs.TaskOverride{
		ContainerOverrides: []*ecs.ContainerOverride{
			containerOverride,
		},
	}

	var params *ecs.RunTaskInput
	if len(t.Subnets) > 0 {
		vpcConfiguration := &ecs.AwsVpcConfiguration{
			AssignPublicIp: aws.String(t.AssignPublicIP),
			Subnets:        t.Subnets,
			SecurityGroups: t.SecurityGroups,
		}
		network := &ecs.NetworkConfiguration{
			AwsvpcConfiguration: vpcConfiguration,
		}
		params = &ecs.RunTaskInput{
			Cluster:              aws.String(t.Cluster),
			TaskDefinition:       taskDefinition.TaskDefinitionArn,
			Overrides:            override,
			NetworkConfiguration: network,
			LaunchType:           aws.String(t.LaunchType),
		}
	} else {
		params = &ecs.RunTaskInput{
			Cluster:        aws.String(t.Cluster),
			TaskDefinition: taskDefinition.TaskDefinitionArn,
			Overrides:      override,
			LaunchType:     aws.String(t.LaunchType),
		}
	}

	resp, err := t.awsECS.RunTaskWithContext(ctx, params)
	if err != nil {
		return nil, err
	}
	if len(resp.Failures) > 0 {
		log.Errorf("Run task error: %+v", resp.Failures)
		return nil, errors.New(*resp.Failures[0].Reason)
	}
	log.Infof("Running tasks: %+v", resp.Tasks)

	err = t.waitRunning(ctx, resp.Tasks)
	if err != nil {
		return resp.Tasks, err
	}
	return resp.Tasks, nil
}

// waitRunning waits a task running.
func (t *Task) waitRunning(ctx context.Context, tasks []*ecs.Task) error {
	log.Info("Waiting for running task...")

	taskArns := []*string{}
	for _, task := range tasks {
		taskArns = append(taskArns, task.TaskArn)
	}
	errCh := make(chan error, 1)
	done := make(chan struct{}, 1)
	go func() {
		err := t.waitExitTasks(taskArns)
		if err != nil {
			errCh <- err
		}
		close(done)
	}()
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-done:
		log.Info("Run task is success")
	case <-ctx.Done():
		return errors.New("process timeout")
	}

	return nil
}

func (t *Task) waitExitTasks(taskArns []*string) error {
retry:
	for {
		time.Sleep(5 * time.Second)

		params := &ecs.DescribeTasksInput{
			Cluster: aws.String(t.Cluster),
			Tasks:   taskArns,
		}
		resp, err := t.awsECS.DescribeTasks(params)
		if err != nil {
			return err
		}

		for _, task := range resp.Tasks {
			if !t.checkTaskStopped(task) {
				continue retry
			}
		}

		for _, task := range resp.Tasks {
			code, result, err := t.checkTaskSucceeded(task)
			if err != nil {
				continue retry
			}
			if !result {
				return errors.Errorf("exit code: %v", code)
			}
		}
		return nil
	}
}

func (t *Task) checkTaskStopped(task *ecs.Task) bool {
	if *task.LastStatus != "STOPPED" {
		return false
	}
	return true
}

func (t *Task) checkTaskSucceeded(task *ecs.Task) (int64, bool, error) {
	for _, c := range task.Containers {
		if c.ExitCode == nil {
			return 1, false, errors.New("can not read exit code")
		}
		if *c.ExitCode != int64(0) {
			return *c.ExitCode, false, nil
		}
	}
	return int64(0), true, nil
}
