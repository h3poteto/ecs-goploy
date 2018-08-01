package cmd

import (
	"fmt"
	"time"

	ecsdeploy "github.com/h3poteto/ecs-goploy/deploy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type runTask struct {
	cluster        string
	name           string
	taskDefinition string
	imageWithTag   string
	command        string
	timeout        int
}

func runTaskCmd() *cobra.Command {
	t := &runTask{}
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Run task on ECS",
		Run:   t.run,
	}

	flags := cmd.Flags()
	flags.StringVarP(&t.cluster, "cluster", "c", "", "Name of ECS cluster")
	flags.StringVarP(&t.name, "container-name", "n", "", "Name of the container for override task definition")
	flags.StringVarP(&t.taskDefinition, "task-definition", "d", "", "Name of task definition to run task. Family and revision (family:revision) or full ARN")
	flags.StringVar(&t.command, "command", "", "Task command which run on ECS")
	flags.IntVarP(&t.timeout, "timeout", "t", 300, "Timeout seconds")

	return cmd
}

func (t *runTask) run(cmd *cobra.Command, args []string) {
	var baseTaskDefinition *string
	if len(t.taskDefinition) > 0 {
		baseTaskDefinition = &t.taskDefinition
	}
	profile, region, verbose := generalConfig()
	if !verbose {
		log.SetLevel(log.ErrorLevel)
	}
	task, err := ecsdeploy.NewTask(t.cluster, t.name, t.command, baseTaskDefinition, (time.Duration(t.timeout) * time.Second), profile, region, verbose)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := task.Run(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Success to run task")
}
