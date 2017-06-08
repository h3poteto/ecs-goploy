/*
Package deploy provides simple functions for deploy ECS.

Usage:

    import "github.com/crowdworks/ecs-goploy/deploy"

Construct a new Deploy, then use deploy functions.

    d := deploy.NewDeploy("cluster", "service-name", "", "", "nginx:stable", 5 * time.Minute, true)

    // deploy new image
    if err := d.Deploy(); err != nil {
        log.Fatalf("[ERROR] %v", err)
    }

Or you can write a custom deploy recipe as you like.

For example:

    d := deploy.NewDeploy("cluster", "service-name", "", "", "nginx:stable", 5 * time.Minute, true)

    // get the current service
    service, err := d.DescribeService()
    if err != nil {
        log.Fatal(err)
    }
    taskDefinition, err := d.DescribeTaskDefinition(service)
    if err != nil {
        log.Fatal(err)
    }
    d.CurrentTask.TaskDefinition = taskDefinition

    newTaskDefinition, err := d.RegisterTaskDefinition(taskDefinition)
    if err != nil {
        log.Fatal(err)
    }
    d.NewTask.TaskDefinition = newTaskDefinition

    // Do something

    err = d.UpdateService(service, newTaskDefinition)
    if err != nil {
        // Do something
    }
    log.Println("[INFO] Deploy success")

*/
package deploy

import (
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/pkg/errors"
)

// Deploy has target ECS information, client of aws-sdk-go, tasks information and timeout seconds.
type Deploy struct {
	awsECS *ecs.ECS

	// Name of ECS cluster.
	Cluster string

	// Name of ECS service.
	Name string

	// Running task information which contains a task definition.
	CurrentTask *Task

	// Task information which will deploy.
	NewTask *Task

	// Wait time when update service.
	// This script monitors ECS service for new task definition to be running after call update service API.
	Timeout time.Duration

	// If deploy failed, rollback to current task definition.
	EnableRollback bool
}

// NewDeploy returns a new Deploy struct, and initialize aws ecs API client.
// Separates imageWithTag into repository and tag, then sets a newTask for deploy.
func NewDeploy(cluster, name, profile, region, imageWithTag string, timeout time.Duration, enableRollback bool) *Deploy {
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
			Image:          image,
			TaskDefinition: nil,
		}
	}
	return &Deploy{
		awsECS,
		cluster,
		name,
		currentTask,
		newTask,
		timeout,
		enableRollback,
	}
}

// Deploy runs deploy commands and handle errors.
func (d *Deploy) Deploy() error {
	service, err := d.DescribeService()
	if err != nil {
		return errors.Wrap(err, "Can not get current service: ")
	}
	taskDefinition, err := d.DescribeTaskDefinition(service)
	if err != nil {
		return errors.Wrap(err, "Can not get current task definition: ")
	}
	d.CurrentTask.TaskDefinition = taskDefinition

	newTaskDefinition, err := d.RegisterTaskDefinition(taskDefinition)
	if err != nil {
		return errors.Wrap(err, "Can not regist new task definition: ")
	}
	d.NewTask.TaskDefinition = newTaskDefinition
	log.Printf("[INFO] new task definition: %+v\n", newTaskDefinition)

	err = d.UpdateService(service, newTaskDefinition)
	if err != nil {
		log.Println("[INFO] update failed")
		updateError := errors.Wrap(err, "Can not update service: ")
		if !d.EnableRollback {
			return updateError
		}
		log.Printf("[INFO] Rolling back to: %+v\n", d.CurrentTask.TaskDefinition)
		if err := d.Rollback(service); err != nil {
			return errors.Wrap(updateError, err.Error())
		}
		return updateError
	}
	return nil
}

// divideImageAndTag separates imageWithTag into repository and tag.
func divideImageAndTag(imageWithTag string) (*string, *string, error) {
	res := strings.Split(imageWithTag, ":")
	if len(res) >= 3 {
		return nil, nil, errors.New("image format is wrong")
	}
	return &res[0], &res[1], nil

}
