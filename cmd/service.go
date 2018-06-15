package cmd

import (
	"log"
	"time"

	ecsdeploy "github.com/h3poteto/ecs-goploy/deploy"
	"github.com/spf13/cobra"
)

type service struct {
	cluster        string
	name           string
	taskDefinition string
	imageWithTag   string
	profile        string
	region         string
	timeout        int
	enableRollback bool
	skipCheckDeployments bool
}

func serviceCmd() *cobra.Command {
	s := &service{}
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Service deploy to ECS",
		Run:   s.service,
	}

	flags := cmd.Flags()
	flags.StringVarP(&s.cluster, "cluster", "c", "", "Name of ECS cluster")
	flags.StringVarP(&s.name, "service-name", "n", "", "Name of service to deploy")
	flags.StringVarP(&s.taskDefinition, "task-definition", "d", "", "Name of base task definition to deploy. Family and revision (family:revision) or full ARN")
	flags.StringVarP(&s.imageWithTag, "image", "i", "", "Name of Docker image to run, ex: repo/image:latest")
	flags.StringVarP(&s.profile, "profile", "p", "", "AWS Profile to use")
	flags.StringVarP(&s.region, "region", "r", "", "AWS Region Name")
	flags.IntVarP(&s.timeout, "timeout", "t", 300, "Timeout seconds. Script monitors ECS Service for new task definition to be running")
	flags.BoolVar(&s.enableRollback, "enable-rollback", false, "Rollback task definition if new version is not running before TIMEOUT")
	flags.BoolVar(&s.skipCheckDeployments, "skip-check-deployments", false, "Skip checking deployments when detect whether deploy completed")

	return cmd
}

func (s *service) service(cmd *cobra.Command, args []string) {
	var baseTaskDefinition *string
	if len(s.taskDefinition) > 0 {
		baseTaskDefinition = &s.taskDefinition
	}
	service, err := ecsdeploy.NewService(s.cluster, s.name, s.imageWithTag, baseTaskDefinition, (time.Duration(s.timeout) * time.Second), s.enableRollback, s.skipCheckDeployments, s.profile, s.region)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	if err := service.Deploy(); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	log.Println("[INFO] Deploy success")
}
