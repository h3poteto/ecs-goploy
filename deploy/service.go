package deploy

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
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
	return d.waitUpdating(5*time.Minute, taskDefinition)
}

// waitUpdating wait new task is deployed.
func (d *Deploy) waitUpdating(timeout time.Duration, newTaskDefinition *ecs.TaskDefinition) error {
	for {
		log.Println("[INFO] Waiting for new task running...")

		time.Sleep(5 * time.Second)

		taskParams := &ecs.ListTasksInput{
			Cluster:     aws.String(d.cluster),
			MaxResults:  aws.Int64(100),
			ServiceName: aws.String(d.name),
		}
		resp, err := d.awsECS.ListTasks(taskParams)
		if err != nil {
			return err
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
			return err
		}
		if d.isNewTaskRunning(currentTasks.Tasks, newTaskDefinition) {
			break
		}
	}

	log.Println("[INFO] New task is running")
	return nil
}

func (d *Deploy) isNewTaskRunning(tasks []*ecs.Task, newTaskDefinition *ecs.TaskDefinition) bool {
	for _, task := range tasks {
		if *task.TaskDefinitionArn == *newTaskDefinition.TaskDefinitionArn {
			if *task.LastStatus == *aws.String("RUNNING") {
				return true
			}
		}
	}
	return false
}
