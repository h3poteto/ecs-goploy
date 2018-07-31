/*
Package deploy provides simple functions for deploy ECS.

Usage:

    import "github.com/h3poteto/ecs-goploy/deploy"

Service update

When you want to update service in ECS, please use this package as follows.

Construct a new Service, then use deploy functions.

    s, err := deploy.NewService("cluster", "service-name", "nginx:stable", nil, 5 * time.Minute, true, "", "")
    if err != nil {
        log.Fatalf("[ERROR] %v", err)
    }

    // deploy new image
    if err := s.Deploy(); err != nil {
        log.Fatalf("[ERROR] %v", err)
    }

Or you can write a custom deploy recipe as you like.

For example:

    s, err := deploy.NewService("cluster", "service-name", "nginx:stable", nil, 5 * time.Minute, true, "", "")
    if err != nil {
        log.Fatal(err)
    }

    // get the current service
    service, err := s.DescribeService()
    if err != nil {
        log.Fatal(err)
    }
    currentTaskDefinition, err := s.TaskDefinition.DescribeTaskDefinition(service)
    if err != nil {
        log.Fatal(err)
    }

    newTaskDefinition, err := s.RegisterTaskDefinition(currentTaskDefinition, s.NewImage)
    if err != nil {
        log.Fatal(err)
    }

    // Do something

    err = s.UpdateService(service, newTaskDefinition)
    if err != nil {
        // Do something
    }
    log.Println("[INFO] Deploy success")

TaskDefinition update

You can create a new revision of the task definition. Please use this task definition at `Task` and `ScheduledTask`.

For example:

    taskDefinition := ecsdeploy.NewTaskDefinition()
    t, err := taskDefinition.Create("sample-task-definition:revision", "nginx:stable")
    if err != nil {
        log.Fatal(err)
    }
    log.Println(*t.TaskDefinitionArn)


Run task

When you want to run task on ECS at once, plese use this package as follows.

For example:

    task, err := ecsdeploy.NewTask("cluster", "container-name", "echo hoge", "sample-task-definition:2", (5 * time.Minute), "", "")
    if err != nil {
        log.Fatal(err)
    }
    if _, err := task.Run(); err != nil {
        log.Fatal(err)
    }
    log.Println("[INFO] Task success")


ScheduledTask update

When you update the ECS Scheduled Task, please use this package.

For example:

    scheduledTask := ecsdeploy.NewScheduledTask()
    scheduledTask("schedule-name", "sample-task-definition:2", 1)

*/
package deploy

import (
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/pkg/errors"
)

// Image has repository and tag string of docker image.
type Image struct {

	// Docker image repository.
	Repository string

	// Docker image tag.
	Tag string
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
	log.Printf("[INFO] New task definition: %+v\n", newTaskDefinition)

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

//Run run task on ECS based on provided task definition.
func (t *Task) Run() ([]*ecs.Task, error) {
	if t.BaseTaskDefinition == nil {
		return nil, errors.New("task definition is required")
	}
	// get a task definition
	baseTaskDefinition, err := t.TaskDefinition.DescribeTaskDefinition(*t.BaseTaskDefinition)
	if err != nil {
		return nil, err
	}

	return t.RunTask(baseTaskDefinition)
}

// Create creates a new revision of the task definition.
func (n *TaskDefinition) Create(base *string, dockerImage string) (*ecs.TaskDefinition, error) {
	repository, revision, err := divideImageAndTag(dockerImage)
	if err != nil {
		return nil, err
	}
	image := &Image{
		Repository: *repository,
		Tag:        *revision,
	}
	if base == nil {
		return nil, errors.New("task definition is required")
	}
	baseTaskDefinition, err := n.DescribeTaskDefinition(*base)
	if err != nil {
		return nil, err
	}
	newTaskDefinition, err := n.RegisterTaskDefinition(baseTaskDefinition, image)
	if err != nil {
		return nil, err
	}
	log.Printf("[INFO] New task definition: %+v\n", newTaskDefinition)

	return newTaskDefinition, nil
}

// Update update the cloudwatch event with provided task definition.
func (s *ScheduledTask) Update(name string, taskDefinition *string, count int64) error {
	if taskDefinition == nil {
		return errors.New("task definition is required")
	}
	// get a task definition
	t, err := s.TaskDefinition.DescribeTaskDefinition(*taskDefinition)
	if err != nil {
		return err
	}

	return s.UpdateTargets(count, t, name)
}
