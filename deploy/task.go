package deploy

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// TaskDefinition get a current task definition
func (d *Deploy) TaskDefinition() error {
	taskArn, err := d.Service()
	if err != nil {
		return err
	}

	params := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(*taskArn),
	}
	resp, err := d.awsECS.DescribeTaskDefinition(params)
	if err != nil {
		return err
	}

	log.Println(resp)
	return nil
}

// Service get target service
func (d *Deploy) Service() (*string, error) {
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

	return resp.Services[0].TaskDefinition, nil
}
