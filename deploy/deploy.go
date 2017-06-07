package deploy

import (
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/pkg/errors"
)

// Deploy have target ecs information
type Deploy struct {
	awsECS      *ecs.ECS
	cluster     string
	name        string
	currentTask *Task
	newTask     *Task
}

// NewDeploy return a new Deploy struct, and initialize aws ecs api client
func NewDeploy(cluster, name, profile, region, imageWithTag string) *Deploy {
	awsECS := ecs.New(session.New(), newConfig(profile, region))
	currentTask := &Task{}
	newTask := &Task{}
	if len(imageWithTag) > 0 {
		var err error
		repository, tag, err := divideImageAndTag(imageWithTag)
		if err != nil {
			log.Fatalf("[ERROR] Can not parse --image parameter: %+v\n", err)
		}
		image := &Image{
			*repository,
			*tag,
		}
		newTask = &Task{
			image:          image,
			taskDefinition: nil,
		}
	}
	return &Deploy{
		awsECS,
		cluster,
		name,
		currentTask,
		newTask,
	}
}

// Deploy run deploy commands
func (d *Deploy) Deploy() error {
	service, err := d.DescribeService()
	if err != nil {
		return errors.Wrap(err, "Can not get current service: ")
	}
	taskDefinition, err := d.TaskDefinition(service)
	if err != nil {
		return errors.Wrap(err, "Can not get current task definition: ")
	}
	d.currentTask.taskDefinition = taskDefinition

	newTaskDefinition, err := d.RegisterTaskDefinition(taskDefinition)
	if err != nil {
		return errors.Wrap(err, "Can not regist new task definition: ")
	}
	d.newTask.taskDefinition = newTaskDefinition
	log.Printf("[INFO] new task definition: %+v\n", newTaskDefinition)

	err = d.UpdateService(service, newTaskDefinition)
	if err != nil {
		log.Println("[INFO] update failed")
		updateError := errors.Wrap(err, "Can not update service: ")
		log.Printf("[INFO] Rolling back to: %+v\n", d.currentTask.taskDefinition)
		if err := d.Rollback(service); err != nil {
			return errors.Wrap(updateError, err.Error())
		}
		return updateError
	}
	return nil
}

func divideImageAndTag(imageWithTag string) (*string, *string, error) {
	res := strings.Split(imageWithTag, ":")
	if len(res) >= 3 {
		return nil, nil, errors.New("image format is wrong.")
	}
	return &res[0], &res[1], nil

}
