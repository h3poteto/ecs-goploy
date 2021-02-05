package cmd

import (
	"fmt"
	"time"

	ecsdeploy "github.com/connectedservices/ecs-goploy/deploy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type runTask struct {
	cluster        string
	name           string
	taskDefinition string
	imageWithTag   string
	command        string
	subnets        string
	securityGroups string
	fargate        bool
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
	flags.StringVarP(&t.subnets, "subnets", "s", "", "Provide subnet IDs with comma-separated string (subnet-12abcde,subnet-34abcde). This param is necessary, if you set farage flag.")
	flags.StringVarP(&t.securityGroups, "security-groups", "g", "", "Provide security group IDs with comma-separated string (sg-0123asdb,sg-2345asdf), if you want to attach the security groups to ENI of the task.")
	flags.BoolVarP(&t.fargate, "fargate", "f", false, "Whether run task with FARGATE")
	flags.IntVarP(&t.timeout, "timeout", "t", 0, "Timeout seconds")

	return cmd
}

func (t *runTask) run(cmd *cobra.Command, args []string) {
	profile, region, verbose := generalConfig()
	if !verbose {
		log.SetLevel(log.ErrorLevel)
	}
	task, err := ecsdeploy.NewTask(t.cluster, t.name, t.command, t.taskDefinition, t.fargate, t.subnets, t.securityGroups, (time.Duration(t.timeout) * time.Second), profile, region, verbose)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := task.Run(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Success to run task")
}
