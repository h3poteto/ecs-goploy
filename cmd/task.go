package cmd

import (
	"log"
	"time"

	ecsdeploy "github.com/crowdworks/ecs-goploy/deploy"
	"github.com/spf13/cobra"
)

type task struct {
	cluster        string
	taskDefinition string
	imageWithTag   string
	profile        string
	region         string
	timeout        int
}

func taskCmd() *cobra.Command {
	t := &task{}
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Run ECS task",
		Run:   t.task,
	}

	flags := cmd.Flags()
	flags.StringVarP(&t.cluster, "cluster", "c", "", "Name of ECS cluster")
	flags.StringVarP(&t.taskDefinition, "task-definition", "d", "", "Name of base task definition to run task. Family and revision (family:revision) or full ARN")
	flags.StringVarP(&t.imageWithTag, "image", "i", "", "Name of Doker image to run, ex: repo/image:latest")
	flags.StringVarP(&t.profile, "profile", "p", "", "AWS Profile to use")
	flags.StringVarP(&t.region, "region", "r", "", "AWS Region Name")
	flags.IntVarP(&t.timeout, "timeout", "t", 300, "Timeout seconds")

	return cmd
}

func (t *task) task(cmd *cobra.Command, args []string) {
	var baseTaskDefinition *string
	if len(t.taskDefinition) > 0 {
		baseTaskDefinition = &t.taskDefinition
	}
	e, err := ecsdeploy.NewDeploy(t.cluster, "", t.profile, t.region, t.imageWithTag, baseTaskDefinition, (time.Duration(t.timeout) * time.Second), false)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	if err := e.Task(); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	log.Println("[INFO] Task success")
}
