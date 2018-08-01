package cmd

import (
	"fmt"

	ecsdeploy "github.com/h3poteto/ecs-goploy/deploy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type updateScheduledTask struct {
	name           string
	taskDefinition string
	count          int64
}

func updateScheduledTaskCmd() *cobra.Command {
	t := &updateScheduledTask{}
	command := &cobra.Command{
		Use:   "scheduled-task",
		Short: "Update ECS Scheduled Task",
		RunE:  t.update,
	}

	flags := command.Flags()
	flags.StringVarP(&t.name, "name", "n", "", "Name of scheduled task")
	flags.StringVarP(&t.taskDefinition, "task-definition", "d", "", "Name of task definition to update scheduled task. Family and revision (family:revision) or full ARN")
	flags.Int64VarP(&t.count, "count", "c", 1, "Count of the task")

	return command
}

func (s *updateScheduledTask) update(cmd *cobra.Command, args []string) error {
	var baseTaskDefinition *string
	if len(s.taskDefinition) > 0 {
		baseTaskDefinition = &s.taskDefinition
	}
	profile, region, verbose := generalConfig()
	if !verbose {
		log.SetLevel(log.ErrorLevel)
	}
	scheduledTask := ecsdeploy.NewScheduledTask(profile, region, verbose)
	err := scheduledTask.Update(s.name, baseTaskDefinition, s.count)
	if err != nil {
		log.Fatal(err)
		return err
	}
	fmt.Println("Success to update the schedule")
	return nil
}
