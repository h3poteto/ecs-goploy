package deploy

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Service has target ECS information, client of aws-sdk-go, tasks information and timeout seconds.
type Service struct {
	awsECS ecsiface.ECSAPI

	// Name of ECS cluster.
	Cluster string

	// Name of ECS service.
	Name string

	// Name of base task definition of deploy.
	BaseTaskDefinition *string

	// TaskDefinition struct to call aws API.
	TaskDefinition *TaskDefinition

	// New image for deploy.
	NewImage *Image

	// Wait time when update service.
	// This script monitors ECS service for new task definition to be running after call update service API.
	Timeout time.Duration

	// If deploy failed, rollback to current task definition.
	EnableRollback bool

	// When check whether deploy completed, confirm only new task status.
	// If this flag is true, confirm service deployments status.
	SkipCheckDeployments bool

	verbose bool
}

// NewService returns a new Service struct, and initialize aws ecs API client.
// Separates imageWithTag into repository and tag, then sets a NewImage for deploy.
func NewService(cluster, name, imageWithTag string, baseTaskDefinition *string, timeout time.Duration, enableRollback bool, skipCheckDeployments bool, profile, region string, verbose bool) (*Service, error) {
	awsECS := ecs.New(session.New(), newConfig(profile, region))
	taskDefinition := NewTaskDefinition(profile, region, verbose)
	if !verbose {
		log.SetLevel(log.ErrorLevel)
	}
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
		skipCheckDeployments,
		verbose,
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
	params := &ecs.UpdateServiceInput{}
	if *service.SchedulingStrategy == "DAEMON" {
		// If the service type is DAEMON, we can not specify desired count.
		params = &ecs.UpdateServiceInput{
			Service:                 aws.String(s.Name),
			Cluster:                 aws.String(s.Cluster),
			DeploymentConfiguration: service.DeploymentConfiguration,
			TaskDefinition:          taskDefinition.TaskDefinitionArn,
		}
	} else {
		params = &ecs.UpdateServiceInput{
			Service:                 aws.String(s.Name),
			Cluster:                 aws.String(s.Cluster),
			DeploymentConfiguration: service.DeploymentConfiguration,
			DesiredCount:            service.DesiredCount,
			TaskDefinition:          taskDefinition.TaskDefinitionArn,
		}
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
	log.Info("Waiting for new task running...")
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
		log.Info("New task is running")
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
		if s.checkCompleteDeploy(service, newTaskDefinition) {
			return nil
		}
	}
}

func (s *Service) checkCompleteDeploy(service *ecs.Service, newTaskDefinition *ecs.TaskDefinition) bool {
	if s.SkipCheckDeployments {
		return s.checkNewTaskRunning(service, newTaskDefinition)
	}
	return s.checkDeployments(service.Deployments, newTaskDefinition)
}

func (s *Service) checkDeployments(deployments []*ecs.Deployment, newTaskDefinition *ecs.TaskDefinition) bool {
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

func (s *Service) checkNewTaskRunning(service *ecs.Service, newTaskDefinition *ecs.TaskDefinition) bool {
	input := &ecs.ListTasksInput{
		Cluster:       service.ClusterArn,
		ServiceName:   service.ServiceName,
		DesiredStatus: aws.String("RUNNING"),
	}
	runningTasks, err := s.awsECS.ListTasks(input)
	if err != nil {
		log.Error(err)
		return false
	}
	params := &ecs.DescribeTasksInput{
		Cluster: service.ClusterArn,
		Tasks:   runningTasks.TaskArns,
	}
	resp, err := s.awsECS.DescribeTasks(params)
	if err != nil {
		log.Error(err)
		return false
	}
	for _, task := range resp.Tasks {
		if *task.LastStatus == "RUNNING" && *task.TaskDefinitionArn == *newTaskDefinition.TaskDefinitionArn {
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
	log.Info("Rolled back")
	return nil
}
