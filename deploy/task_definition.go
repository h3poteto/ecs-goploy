package deploy

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// TaskDefinition has image and task definition information.
type TaskDefinition struct {
	awsECS *ecs.ECS
}

func NewTaskDefinition(profile, region string) *TaskDefinition {
	awsECS := ecs.New(session.New(), newConfig(profile, region))
	return &TaskDefinition{
		awsECS,
	}
}

// DescribeTaskDefinition gets a task definition.
// The family for the latest ACTIVE revision, family and revision (family:revision)
// for a specific revision in the family, or full Amazon Resource Name (ARN)
// of the task definition to describe.
func (d *TaskDefinition) DescribeTaskDefinition(taskDefinitionName string) (*ecs.TaskDefinition, error) {
	params := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinitionName),
	}
	resp, err := d.awsECS.DescribeTaskDefinition(params)
	if err != nil {
		return nil, err
	}

	return resp.TaskDefinition, nil
}

// RegisterTaskDefinition registers new task definition if needed.
// If newTask is not set, returns a task definition which same as the given task definition.
func (d *TaskDefinition) RegisterTaskDefinition(baseDefinition *ecs.TaskDefinition, newImage *Image) (*ecs.TaskDefinition, error) {
	var containerDefinitions []*ecs.ContainerDefinition
	for _, c := range baseDefinition.ContainerDefinitions {
		newDefinition, err := d.NewContainerDefinition(c, newImage)
		if err != nil {
			return nil, err
		}
		containerDefinitions = append(containerDefinitions, newDefinition)
	}
	params := &ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: containerDefinitions,
		Family:               baseDefinition.Family,
		NetworkMode:          baseDefinition.NetworkMode,
		PlacementConstraints: baseDefinition.PlacementConstraints,
		TaskRoleArn:          baseDefinition.TaskRoleArn,
		Volumes:              baseDefinition.Volumes,
	}

	resp, err := d.awsECS.RegisterTaskDefinition(params)
	if err != nil {
		return nil, err
	}

	return resp.TaskDefinition, nil
}

// NewContainerDefinition updates image tag in the given container definition.
// If the container definition is not target container, returns the givien definition.
func (d *TaskDefinition) NewContainerDefinition(baseDefinition *ecs.ContainerDefinition, newImage *Image) (*ecs.ContainerDefinition, error) {
	if newImage == nil {
		return baseDefinition, nil
	}
	baseRepository, _, err := divideImageAndTag(*baseDefinition.Image)
	if err != nil {
		return nil, err
	}
	if newImage.Repository != *baseRepository {
		return baseDefinition, nil
	}
	imageWithTag := (newImage.Repository) + ":" + (newImage.Tag)
	baseDefinition.Image = &imageWithTag
	return baseDefinition, nil
}
