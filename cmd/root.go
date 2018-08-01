package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd root comand for cobra
var RootCmd = &cobra.Command{
	Use:           "ecs-goploy",
	Short:         "Deploy commands for ECS",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	cobra.OnInitialize()
	RootCmd.PersistentFlags().StringP("profile", "", "", "AWS profile (detault is none, and use environment variables)")
	RootCmd.PersistentFlags().StringP("region", "", "", "AWS region (default is none, and use AWS_DEFAULT_REGION)")
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose mode")
	viper.BindPFlag("profile", RootCmd.PersistentFlags().Lookup("profile"))
	viper.BindPFlag("region", RootCmd.PersistentFlags().Lookup("region"))
	viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))

	RootCmd.AddCommand(
		versionCmd(),
		runCmd(),
		updateCmd(),
	)
}

// generalConfig returns profile, and region.
func generalConfig() (string, string, bool) {
	return viper.GetString("profile"), viper.GetString("region"), viper.GetBool("verbose")
}
