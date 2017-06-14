/*
Package deploy provides simple functions for deploy ECS.

Usage:

    import "github.com/crowdworks/ecs-goploy/deploy"

Construct a new Deploy, then use deploy functions.

    d, err := deploy.NewDeploy("cluster", "service-name", "", "", "nginx:stable", nil, 5 * time.Minute, true)
    if err != nil {
        log.Fatalf("[ERROR] %v", err)
    }

    // deploy new image
    if err := d.Deploy(); err != nil {
        log.Fatalf("[ERROR] %v", err)
    }

Or you can write a custom deploy recipe as you like.

For example:

    d, err := deploy.NewDeploy("cluster", "service-name", "", "", "nginx:stable", nil, 5 * time.Minute, true)
    if err != nil {
        log.Fatal(err)
    }

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

	"github.com/pkg/errors"
)

// Image has repository and tag string.
type Image struct {
	Repository string
	Tag        string
}

// Deploy runs deploy commands and handle errors.
func (s *Service) Deploy() error {
	service, err := s.DescribeService()
	if err != nil {
		return errors.Wrap(err, "Can not get current service: ")
	}

	// get running task definition
	currentTaskDefinition, err := s.TaskDefinition.DescribeTaskDefinition(*service.TaskDefinition)
	if err != nil {
		return errors.Wrap(err, "Can not get task definition: ")
	}

	// get base task definition if needed
	baseTaskDefinition := currentTaskDefinition
	if s.BaseTaskDefinition != nil {
		var err error
		baseTaskDefinition, err = s.TaskDefinition.DescribeTaskDefinition(*s.BaseTaskDefinition)
		if err != nil {
			return errors.Wrap(err, "Can not get task definition: ")
		}
	}

	newTaskDefinition, err := s.TaskDefinition.RegisterTaskDefinition(baseTaskDefinition, s.NewImage)
	if err != nil {
		return errors.Wrap(err, "Can not regist new task definition: ")
	}
	log.Printf("[INFO] new task definition: %+v\n", newTaskDefinition)

	err = s.UpdateService(service, newTaskDefinition)
	if err != nil {
		log.Println("[INFO] update failed")
		updateError := errors.Wrap(err, "Can not update service: ")
		if !s.EnableRollback {
			return updateError
		}

		// rollback to the current task definition which have been running to the end
		log.Printf("[INFO] Rolling back to: %+v\n", currentTaskDefinition)
		if err := s.Rollback(service, currentTaskDefinition); err != nil {
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

//Run regists a new task definition and run task on ECS.
func (t *Task) Run() error {
	// get a task definition
	baseTaskDefinition, err := t.TaskDefinition.DescribeTaskDefinition(*t.BaseTaskDefinition)
	if err != nil {
		return err
	}
	// add new task definition to run task
	newTaskDefinition, err := t.TaskDefinition.RegisterTaskDefinition(baseTaskDefinition, t.NewImage)
	if err != nil {
		return nil
	}
	log.Printf("[INFO] New task definition: %+v\n", newTaskDefinition)

	if err := t.RunTask(newTaskDefinition, t.Timeout); err != nil {
		return err
	}
	return nil
}
