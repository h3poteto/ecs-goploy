package deploy

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/pkg/errors"
)

// DescribeService get current service
func (d *Deploy) DescribeService() (*ecs.Service, error) {
	params := &ecs.DescribeServicesInput{
		Services: []*string{
			aws.String(d.name),
		},
		Cluster: aws.String(d.cluster),
	}
	resp, err := d.awsECS.DescribeServices(params)
	if err != nil {
		return nil, err
	}

	return resp.Services[0], nil
}

// UpdateService update a service with a new task definition, and wait during update action
func (d *Deploy) UpdateService(service *ecs.Service, taskDefinition *ecs.TaskDefinition) error {
	params := &ecs.UpdateServiceInput{
		Service:                 aws.String(d.name),
		Cluster:                 aws.String(d.cluster),
		DeploymentConfiguration: service.DeploymentConfiguration,
		DesiredCount:            service.DesiredCount,
		TaskDefinition:          taskDefinition.TaskDefinitionArn,
	}
	resp, err := d.awsECS.UpdateService(params)
	if err != nil {
		return err
	}
	log.Println(resp)
	newService := resp.Service
	if *newService.DesiredCount <= 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	return d.waitUpdating(ctx, taskDefinition)
}

// waitUpdating wait new task is deployed.
func (d *Deploy) waitUpdating(ctx context.Context, newTaskDefinition *ecs.TaskDefinition) error {
	log.Println("[INFO] Waiting for new task start...")

	newTask, err := d.waitNewTaskStart(ctx, newTaskDefinition)
	if err != nil {
		return err
	}

	log.Println("[INFO] Waiting for new task is available...")

	params := &ecs.DescribeTasksInput{
		Cluster: aws.String(d.cluster),
		Tasks: []*string{
			aws.String(*newTask.TaskArn),
		},
	}
	err = d.awsECS.WaitUntilTasksRunningWithContext(ctx, params)
	if err != nil {
		return err
	}

	log.Println("[INFO] New task is available")
	return nil
}

func (d *Deploy) waitNewTaskStart(ctx context.Context, newTaskDefinition *ecs.TaskDefinition) (*ecs.Task, error) {
	newTaskCh := make(chan *ecs.Task, 1)
	errCh := make(chan error, 1)

	go func() {
		for {
			time.Sleep(5 * time.Second)

			taskParams := &ecs.ListTasksInput{
				Cluster:     aws.String(d.cluster),
				MaxResults:  aws.Int64(100),
				ServiceName: aws.String(d.name),
			}
			resp, err := d.awsECS.ListTasks(taskParams)
			if err != nil {
				errCh <- err
				return
			}
			currentTaskArns := resp.TaskArns

			if len(currentTaskArns) <= 0 {
				continue
			}
			params := &ecs.DescribeTasksInput{
				Cluster: aws.String(d.cluster),
				Tasks:   currentTaskArns,
			}
			currentTasks, err := d.awsECS.DescribeTasks(params)
			if err != nil {
				errCh <- err
				return
			}
			task := d.findNewTask(currentTasks.Tasks, newTaskDefinition)
			if task != nil {
				newTaskCh <- task
				return
			}
		}
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
	case newTask := <-newTaskCh:
		return newTask, nil
	case <-ctx.Done():
		return nil, errors.New("process timeout")
	}
	return nil, errors.New("can not find new task")
}

func (d *Deploy) findNewTask(tasks []*ecs.Task, newTaskDefinition *ecs.TaskDefinition) *ecs.Task {
	for _, task := range tasks {
		if *task.TaskDefinitionArn == *newTaskDefinition.TaskDefinitionArn {
			return task
		}
	}
	return nil
}
