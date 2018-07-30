package cmd

import (
	"log"

	ecsdeploy "github.com/h3poteto/ecs-goploy/deploy"
	"github.com/spf13/cobra"
)

type updateTaskDefinition struct {
	baseTaskDefinition string
	imageWithTag       string
	profile            string
	region             string
}

func updateTaskDefinitionCmd() *cobra.Command {
	n := &updateTaskDefinition{}
	cmd := &cobra.Command{
		Use:   "task-definition",
		Short: "Create a new revision of the task definiition",
		RunE:  n.create,
	}

	flags := cmd.Flags()
	flags.StringVarP(&n.baseTaskDefinition, "base-task-definition", "d", "", "Nmae of base task definition to create a new revision. Family and revision (family:revision) or full ARN")
	flags.StringVarP(&n.imageWithTag, "image", "i", "", "Name of Docker image to update, ex: repo/image:latest")
	flags.StringVarP(&n.profile, "profile", "p", "", "AWS Profile to use")
	flags.StringVarP(&n.region, "region", "r", "", "AWS Region Name")

	return cmd
}

func (n *updateTaskDefinition) create(cmd *cobra.Command, args []string) error {
	var baseTaskDefinition *string
	if len(n.baseTaskDefinition) > 0 {
		baseTaskDefinition = &n.baseTaskDefinition
	}
	taskDefinition := ecsdeploy.NewTaskDefinition(n.profile, n.region)
	t, err := taskDefinition.Create(baseTaskDefinition, n.imageWithTag)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
		return err
	}
	log.Println(*t.TaskDefinitionArn)
	return nil
}
