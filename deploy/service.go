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

	newService := resp.Service
	if *newService.DesiredCount <= 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	return d.waitUpdating(ctx, taskDefinition)
}

// waitUpdating wait new task is deployed.
func (d *Deploy) waitUpdating(ctx context.Context, newTaskDefinition *ecs.TaskDefinition) error {
	log.Println("[INFO] Waiting for new task running...")
	errCh := make(chan error, 1)
	done := make(chan struct{}, 1)
	go func() {
		err := d.waitSwitchTask(newTaskDefinition)
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
		log.Println("[INFO] New task is running, and old task is stopped")
	case <-ctx.Done():
		return errors.New("process timeout")
	}

	return nil
}

func (d *Deploy) waitSwitchTask(newTaskDefinition *ecs.TaskDefinition) error {
	for {
		time.Sleep(5 * time.Second)

		service, err := d.DescribeService()
		if err != nil {
			return err
		}
		if d.checkNewTaskRunning(service.Deployments, newTaskDefinition) {
			return nil
		}
	}
}

func (d *Deploy) checkNewTaskRunning(deployments []*ecs.Deployment, newTaskDefinition *ecs.TaskDefinition) bool {
	if len(deployments) != 1 {
		return false
	}
	for _, deploy := range deployments {
		if *deploy.TaskDefinition == *newTaskDefinition.TaskDefinitionArn && *deploy.Status == "PRIMARY" && *deploy.DesiredCount == *deploy.RunningCount {
			return true
		}
	}
	return false
}

func (d *Deploy) Rollback(service *ecs.Service) error {
	if d.currentTask == nil || d.currentTask.taskDefinition == nil {
		return errors.New("old task definition is not exist")
	}
	params := &ecs.UpdateServiceInput{
		Service:                 aws.String(d.name),
		Cluster:                 aws.String(d.cluster),
		DeploymentConfiguration: service.DeploymentConfiguration,
		DesiredCount:            service.DesiredCount,
		TaskDefinition:          d.currentTask.taskDefinition.TaskDefinitionArn,
	}
	_, err := d.awsECS.UpdateService(params)
	if err != nil {
		return err
	}
	log.Println("[INFO] Rolled back")
	return nil
}
