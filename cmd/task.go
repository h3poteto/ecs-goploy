package cmd

import (
	"log"
	"time"

	ecsdeploy "github.com/h3poteto/ecs-goploy/deploy"
	"github.com/spf13/cobra"
)

type task struct {
	cluster        string
	name           string
	taskDefinition string
	imageWithTag   string
	profile        string
	region         string
	command        string
	timeout        int
}

func taskCmd() *cobra.Command {
	t := &task{}
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Run task on ECS",
		Run:   t.task,
	}

	flags := cmd.Flags()
	flags.StringVarP(&t.cluster, "cluster", "c", "", "Name of ECS cluster")
	flags.StringVarP(&t.name, "container-name", "n", "", "Name of the container for override task definition")
	flags.StringVarP(&t.taskDefinition, "task-definition", "d", "", "Name of task definition to run task. Family and revision (family:revision) or full ARN")
	flags.StringVarP(&t.imageWithTag, "image", "i", "", "Name of Doker image to run, ex: repo/image:latest")
	flags.StringVarP(&t.profile, "profile", "p", "", "AWS Profile to use")
	flags.StringVarP(&t.region, "region", "r", "", "AWS Region Name")
	flags.StringVar(&t.command, "command", "", "Task command which run on ECS")
	flags.IntVarP(&t.timeout, "timeout", "t", 300, "Timeout seconds")

	return cmd
}

func (t *task) task(cmd *cobra.Command, args []string) {
	var baseTaskDefinition *string
	if len(t.taskDefinition) > 0 {
		baseTaskDefinition = &t.taskDefinition
	}
	task, err := ecsdeploy.NewTask(t.cluster, t.name, t.imageWithTag, t.command, baseTaskDefinition, (time.Duration(t.timeout) * time.Second), t.profile, t.region)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	if _, err := task.Run(); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	log.Println("[INFO] Task success")
}
