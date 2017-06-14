package deploy

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// TaskDefinition has image and task definition information.
type TaskDefinition struct {
	Image          *Image
	TaskDefinition *ecs.TaskDefinition
}

// Image has repository and tag string.
type Image struct {
	Repository string
	Tag        string
}

// DescribeTaskDefinition gets a task definition.
// The family for the latest ACTIVE revision, family and revision (family:revision)
// for a specific revision in the family, or full Amazon Resource Name (ARN)
// of the task definition to describe.
func (d *Deploy) DescribeTaskDefinition(taskDefinitionName string) (*ecs.TaskDefinition, error) {
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
func (d *Deploy) RegisterTaskDefinition(baseDefinition *ecs.TaskDefinition) (*ecs.TaskDefinition, error) {
	var containerDefinitions []*ecs.ContainerDefinition
	for _, c := range baseDefinition.ContainerDefinitions {
		newDefinition, err := d.NewContainerDefinition(c)
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
func (d *Deploy) NewContainerDefinition(baseDefinition *ecs.ContainerDefinition) (*ecs.ContainerDefinition, error) {
	if d.NewTask.Image == nil {
		return baseDefinition, nil
	}
	baseRepository, _, err := divideImageAndTag(*baseDefinition.Image)
	if err != nil {
		return nil, err
	}
	if d.NewTask.Image.Repository != *baseRepository {
		return baseDefinition, nil
	}
	imageWithTag := (d.NewTask.Image.Repository) + ":" + (d.NewTask.Image.Tag)
	baseDefinition.Image = &imageWithTag
	return baseDefinition, nil
}
