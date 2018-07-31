package cmd

import (
	ecsdeploy "github.com/h3poteto/ecs-goploy/deploy"
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
	profile, region := generalConfig()
	scheduledTask := ecsdeploy.NewScheduledTask(profile, region)
	return scheduledTask.Update(s.name, baseTaskDefinition, s.count)
}
