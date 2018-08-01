package cmd

import (
	"fmt"

	ecsdeploy "github.com/h3poteto/ecs-goploy/deploy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type updateTaskDefinition struct {
	baseTaskDefinition string
	imageWithTag       string
}

func updateTaskDefinitionCmd() *cobra.Command {
	n := &updateTaskDefinition{}
	cmd := &cobra.Command{
		Use:   "task-definition",
		Short: "Create a new revision of the task definiition",
		RunE:  n.update,
	}

	flags := cmd.Flags()
	flags.StringVarP(&n.baseTaskDefinition, "base-task-definition", "d", "", "Nmae of base task definition to create a new revision. Family and revision (family:revision) or full ARN")
	flags.StringVarP(&n.imageWithTag, "image", "i", "", "Name of Docker image to update, ex: repo/image:latest")

	return cmd
}

func (n *updateTaskDefinition) update(cmd *cobra.Command, args []string) error {
	var baseTaskDefinition *string
	if len(n.baseTaskDefinition) > 0 {
		baseTaskDefinition = &n.baseTaskDefinition
	}
	profile, region, verbose := generalConfig()
	if !verbose {
		log.SetLevel(log.ErrorLevel)
	}
	taskDefinition := ecsdeploy.NewTaskDefinition(profile, region, verbose)
	t, err := taskDefinition.Create(baseTaskDefinition, n.imageWithTag)
	if err != nil {
		log.Fatal(err)
		return err
	}
	fmt.Println(*t.TaskDefinitionArn)
	return nil
}
