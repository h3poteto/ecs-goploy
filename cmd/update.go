package cmd

import "github.com/spf13/cobra"

func updateCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "update",
		Short: "Update some ECS resource",
		Run: func(c *cobra.Command, arg []string) {
			c.Help()
		},
	}

	command.AddCommand(
		updateServiceCmd(),
		updateTaskDefinitionCmd(),
		updateScheduledTaskCmd(),
	)

	return command
}
