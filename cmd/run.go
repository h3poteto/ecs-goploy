package cmd

import "github.com/spf13/cobra"

func runCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "run",
		Short: "Run command",
		Run: func(c *cobra.Command, arg []string) {
			c.Help()
		},
	}
	command.AddCommand(
		runTaskCmd(),
	)

	return command
}
