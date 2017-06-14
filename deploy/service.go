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

// Service has target ECS information, client of aws-sdk-go, tasks information and timeout seconds.
type Service struct {
	awsECS *ecs.ECS

	// Name of ECS cluster.
	Cluster string

	// Name of ECS service.
	Name string

	// Name of base task definition of deploy.
	BaseTaskDefinition *string

	TaskDefinition *TaskDefinition

	NewImage *Image

	// Wait time when update service.
	// This script monitors ECS service for new task definition to be running after call update service API.
	Timeout time.Duration

	// If deploy failed, rollback to current task definition.
	EnableRollback bool
}

// NewService returns a new Service struct, and initialize aws ecs API client.
// Separates imageWithTag into repository and tag, then sets a newTask for deploy.
func NewService(cluster, name, imageWithTag string, baseTaskDefinition *string, timeout time.Duration, enableRollback bool, profile, region string) (*Service, error) {
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
	return &Service{
		awsECS,
		cluster,
		name,
		baseTaskDefinition,
		taskDefinition,
		newImage,
		timeout,
		enableRollback,
	}, nil
}

// DescribeService gets a current service in the cluster.
func (s *Service) DescribeService() (*ecs.Service, error) {
	params := &ecs.DescribeServicesInput{
		Services: []*string{
			aws.String(s.Name),
		},
		Cluster: aws.String(s.Cluster),
	}
	resp, err := s.awsECS.DescribeServices(params)
	if err != nil {
		return nil, err
	}

	return resp.Services[0], nil
}

// UpdateService updates the service with a new task definition, and wait during update action.
func (s *Service) UpdateService(service *ecs.Service, taskDefinition *ecs.TaskDefinition) error {
	params := &ecs.UpdateServiceInput{
		Service:                 aws.String(s.Name),
		Cluster:                 aws.String(s.Cluster),
		DeploymentConfiguration: service.DeploymentConfiguration,
		DesiredCount:            service.DesiredCount,
		TaskDefinition:          taskDefinition.TaskDefinitionArn,
	}
	resp, err := s.awsECS.UpdateService(params)
	if err != nil {
		return err
	}

	newService := resp.Service
	if *newService.DesiredCount <= 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()
	return s.waitUpdating(ctx, taskDefinition)
}

// waitUpdating waits the new task definition is deployed.
func (s *Service) waitUpdating(ctx context.Context, newTaskDefinition *ecs.TaskDefinition) error {
	log.Println("[INFO] Waiting for new task running...")
	errCh := make(chan error, 1)
	done := make(chan struct{}, 1)
	go func() {
		err := s.waitSwitchTask(newTaskDefinition)
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

func (s *Service) waitSwitchTask(newTaskDefinition *ecs.TaskDefinition) error {
	for {
		time.Sleep(5 * time.Second)

		service, err := s.DescribeService()
		if err != nil {
			return err
		}
		if s.checkNewTaskRunning(service.Deployments, newTaskDefinition) {
			return nil
		}
	}
}

func (s *Service) checkNewTaskRunning(deployments []*ecs.Deployment, newTaskDefinition *ecs.TaskDefinition) bool {
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

// Rollback updates the service with current task definition.
// This method call update-service API and does not wait for execution to end.
func (s *Service) Rollback(service *ecs.Service, currentTaskDefinition *ecs.TaskDefinition) error {
	if currentTaskDefinition == nil {
		return errors.New("old task definition is not exist")
	}
	params := &ecs.UpdateServiceInput{
		Service:                 aws.String(s.Name),
		Cluster:                 aws.String(s.Cluster),
		DeploymentConfiguration: service.DeploymentConfiguration,
		DesiredCount:            service.DesiredCount,
		TaskDefinition:          currentTaskDefinition.TaskDefinitionArn,
	}
	_, err := s.awsECS.UpdateService(params)
	if err != nil {
		return err
	}
	log.Println("[INFO] Rolled back")
	return nil
}
