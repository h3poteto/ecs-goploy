package cmd

import (
	"log"

	ecsdeploy "github.com/crowdworks/ecs-goploy/deploy"
	"github.com/spf13/cobra"
)

type deploy struct {
	cluster      string
	name         string
	imageWithTag string
	profile      string
	region       string
}

func deployCmd() *cobra.Command {
	d := &deploy{}
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy ecs",
		Run:   d.deploy,
	}

	flags := cmd.Flags()
	flags.StringVarP(&d.cluster, "cluster", "c", "", "Name of ECS cluster")
	flags.StringVarP(&d.name, "service-name", "n", "", "Name of service to deploy")
	flags.StringVarP(&d.imageWithTag, "image", "i", "", "Name of Docker image to run, ex: repo/image:latest")
	flags.StringVarP(&d.profile, "profile", "p", "", "AWS Profile to use")
	flags.StringVarP(&d.region, "region", "r", "", "AWS Region Name")

	return cmd
}

func (d *deploy) deploy(cmd *cobra.Command, args []string) {
	e := ecsdeploy.NewDeploy(d.cluster, d.name, d.profile, d.region, d.imageWithTag)
	if err := e.Deploy(); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	log.Println("[INFO] Deploy success")
}
