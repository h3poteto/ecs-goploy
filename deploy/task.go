package deploy

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/pkg/errors"
)

type Task struct {
	awsECS *ecs.ECS

	Cluster string

	BaseTaskDefinition *string

	TaskDefinition *TaskDefinition

	NewImage *Image

	Timeout time.Duration
}

func NewTask(cluster, imageWithTag string, baseTaskDefinition *string, timeout time.Duration, profile, region string) (*Task, error) {
	if baseTaskDefinition == nil {
		return nil, errors.New("task definition is required")
	}
	awsECS := ecs.New(session.New(), newConfig(profile, region))
	taskDefinition := NewTaskDefinition(profile, region)
	var newImage *Image
	if len(imageWithTag) > 0 {
		var err error
		repository, tag, err := divideImageAndTag(imageWithTag)
		if err != nil {
			return nil, err
		}
		newImage = &Image{
			*repository,
			*tag,
		}
	}
	return &Task{
		awsECS,
		cluster,
		baseTaskDefinition,
		taskDefinition,
		newImage,
		timeout,
	}, nil
}

// RunTask calls run-task API.
func (t *Task) RunTask(taskDefinition *ecs.TaskDefinition, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	params := &ecs.RunTaskInput{
		Cluster:        aws.String(t.Cluster),
		TaskDefinition: taskDefinition.TaskDefinitionArn,
	}
	resp, err := t.awsECS.RunTaskWithContext(ctx, params)
	if err != nil {
		return err
	}

	return t.waitRunning(ctx, resp.Tasks)
}

// waitRunning waits a task running.
func (t *Task) waitRunning(ctx context.Context, tasks []*ecs.Task) error {
	log.Println("[INFO] Waiting for running task...")

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
		log.Println("[INFO] Run task is success")
	case <-ctx.Done():
		return errors.New("process timeout")
	}

	return nil
}

func (t *Task) waitExitTasks(taskArns []*string) error {
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
				continue
			}
		}

		for _, task := range resp.Tasks {
			if !t.checkTaskSucceeded(task) {
				return errors.New("exit code is not zero")
			}
		}
		return nil
	}
}

func (t *Task) checkTaskStopped(task *ecs.Task) bool {
	if *task.DesiredStatus != "STOPPED" {
		return false
	}
	return true
}

func (t *Task) checkTaskSucceeded(task *ecs.Task) bool {
	for _, c := range task.Containers {
		if *c.ExitCode != int64(0) {
			return false
		}
	}
	return true
}
