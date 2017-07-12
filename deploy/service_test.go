package deploy

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
)

type mockedDescribeServices struct {
	ecsiface.ECSAPI
	Resp ecs.DescribeServicesOutput
}

func (m mockedDescribeServices) DescribeServices(in *ecs.DescribeServicesInput) (*ecs.DescribeServicesOutput, error) {
	return &m.Resp, nil
}

func TestDescribeService(t *testing.T) {
	s := &ecs.Service{
		ServiceName: aws.String("dummy-service"),
	}
	resp := ecs.DescribeServicesOutput{
		Services: []*ecs.Service{
			s,
		},
	}
	service := &Service{
		awsECS: mockedDescribeServices{Resp: resp},
	}
	output, err := service.DescribeService()
	if err != nil {
		t.Error(err)
	}

	if *output.ServiceName != "dummy-service" {
		t.Error("ServiceName is invalid")
	}

}

type mockedUpdateService struct {
	ecsiface.ECSAPI
	Update   ecs.UpdateServiceOutput
	Describe ecs.DescribeServicesOutput
}

func (m mockedUpdateService) UpdateService(in *ecs.UpdateServiceInput) (*ecs.UpdateServiceOutput, error) {
	return &m.Update, nil
}

func (m mockedUpdateService) DescribeServices(in *ecs.DescribeServicesInput) (*ecs.DescribeServicesOutput, error) {
	return &m.Describe, nil
}

func TestUpdateService(t *testing.T) {
	s := &ecs.Service{
		ServiceName: aws.String("dummy-service"),
	}
	newTaskDefinition := &ecs.TaskDefinition{
		TaskDefinitionArn: aws.String("task-definition-arn"),
	}
	describe := ecs.DescribeServicesOutput{
		Services: []*ecs.Service{
			&ecs.Service{
				ServiceName: aws.String("dummy-service"),
				Deployments: []*ecs.Deployment{
					&ecs.Deployment{
						TaskDefinition: aws.String("task-definition-arn"),
						Status:         aws.String("PRIMARY"),
						DesiredCount:   aws.Int64(1),
						RunningCount:   aws.Int64(1),
					},
				},
			},
		},
	}
	update := ecs.UpdateServiceOutput{
		Service: &ecs.Service{
			ServiceName:  aws.String("dummy-service"),
			DesiredCount: aws.Int64(1),
		},
	}

	service := &Service{
		awsECS: mockedUpdateService{
			Update:   update,
			Describe: describe,
		},
		Timeout: 10 * time.Second,
	}

	err := service.UpdateService(s, newTaskDefinition)
	if err != nil {
		t.Error(err)
	}
}
