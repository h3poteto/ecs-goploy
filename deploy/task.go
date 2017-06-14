package deploy

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/pkg/errors"
)

// RunTask calls run-task API.
func (d *Deploy) RunTask(taskDefinition *ecs.TaskDefinition, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	params := &ecs.RunTaskInput{
		Cluster:        aws.String(d.Cluster),
		TaskDefinition: taskDefinition.TaskDefinitionArn,
	}
	resp, err := d.awsECS.RunTaskWithContext(ctx, params)
	if err != nil {
		return err
	}

	return d.waitRunning(ctx, resp.Tasks)
}

// waitRunning waits a task running.
func (d *Deploy) waitRunning(ctx context.Context, tasks []*ecs.Task) error {
	log.Println("[INFO] Waiting for running task...")

	taskArns := []*string{}
	for _, t := range tasks {
		taskArns = append(taskArns, t.TaskArn)
	}
	errCh := make(chan error, 1)
	done := make(chan struct{}, 1)
	go func() {
		err := d.waitExitTasks(taskArns)
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

func (d *Deploy) waitExitTasks(taskArns []*string) error {
	for {
		time.Sleep(5 * time.Second)

		params := &ecs.DescribeTasksInput{
			Cluster: aws.String(d.Cluster),
			Tasks:   taskArns,
		}
		resp, err := d.awsECS.DescribeTasks(params)
		if err != nil {
			return err
		}

		for _, t := range resp.Tasks {
			if !d.checkTaskStopped(t) {
				continue
			}
		}

		for _, t := range resp.Tasks {
			if !d.checkTaskSucceeded(t) {
				return errors.New("exit code is not zero")
			}
		}
		return nil
	}
}

func (d *Deploy) checkTaskStopped(task *ecs.Task) bool {
	if *task.DesiredStatus != "STOPPED" {
		return false
	}
	return true
}

func (d *Deploy) checkTaskSucceeded(task *ecs.Task) bool {
	for _, c := range task.Containers {
		if *c.ExitCode != int64(0) {
			return false
		}
	}
	return true
}
