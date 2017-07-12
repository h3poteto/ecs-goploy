package deploy

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
)

type mockedDescribeTaskDefinition struct {
	ecsiface.ECSAPI
	Resp ecs.DescribeTaskDefinitionOutput
}

func (m mockedDescribeTaskDefinition) DescribeTaskDefinition(in *ecs.DescribeTaskDefinitionInput) (*ecs.DescribeTaskDefinitionOutput, error) {
	return &m.Resp, nil
}

func TestDescribeTaskDefinition(t *testing.T) {
	resp := ecs.DescribeTaskDefinitionOutput{
		TaskDefinition: &ecs.TaskDefinition{
			Family: aws.String("dummy"),
		},
	}
	taskDefinition := &TaskDefinition{
		awsECS: mockedDescribeTaskDefinition{Resp: resp},
	}
	output, err := taskDefinition.DescribeTaskDefinition("dummy")
	if err != nil {
		t.Error(err)
	}

	if *output.Family != "dummy" {
		t.Error("Task definition is invalid")
	}

}

type mockedRegisterTaskDefinition struct {
	ecsiface.ECSAPI
	Resp ecs.RegisterTaskDefinitionOutput
}

func (m mockedRegisterTaskDefinition) RegisterTaskDefinition(in *ecs.RegisterTaskDefinitionInput) (*ecs.RegisterTaskDefinitionOutput, error) {
	return &m.Resp, nil
}

func TestRegisterTaskDefinition(t *testing.T) {
	resp := ecs.RegisterTaskDefinitionOutput{
		TaskDefinition: &ecs.TaskDefinition{
			Family: aws.String("dummy"),
		},
	}

	taskDefinition := &TaskDefinition{
		awsECS: mockedRegisterTaskDefinition{Resp: resp},
	}
	output, err := taskDefinition.RegisterTaskDefinition(
		&ecs.TaskDefinition{
			Family: aws.String("dummy"),
		},
		&Image{
			Repository: "nginx",
			Tag:        "latest",
		},
	)
	if err != nil {
		t.Error(err)
	}
	if *output.Family != "dummy" {
		t.Error("Task definition is invalid")
	}
}

type mockedECS struct {
	ecsiface.ECSAPI
}

func TestNewContainerDefinition(t *testing.T) {
	baseDefinition := &ecs.ContainerDefinition{
		Image: aws.String("nginx:latest"),
	}
	newImage := &Image{
		Repository: "nginx",
		Tag:        "master",
	}
	taskDefinition := &TaskDefinition{
		awsECS: mockedECS{},
	}
	newContainer, err := taskDefinition.NewContainerDefinition(baseDefinition, newImage)
	if err != nil {
		t.Error(err)
	}
	if *newContainer.Image != "nginx:master" {
		t.Error("Container definition is invalid")
	}
}
