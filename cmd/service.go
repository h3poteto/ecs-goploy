package cmd

import (
	"fmt"
	"time"

	ecsdeploy "github.com/connectedservices/ecs-goploy/deploy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type updateService struct {
	cluster              string
	name                 string
	baseTaskDefinition   string
	imageWithTag         string
	timeout              int
	enableRollback       bool
	skipCheckDeployments bool
}

func updateServiceCmd() *cobra.Command {
	s := &updateService{}
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Deploy an ECS Service",
		Run:   s.update,
	}

	flags := cmd.Flags()
	flags.StringVarP(&s.cluster, "cluster", "c", "", "Name of ECS cluster")
	flags.StringVarP(&s.name, "service-name", "n", "", "Name of service to deploy")
	flags.StringVarP(&s.baseTaskDefinition, "base-task-definition", "d", "", "Name of base task definition to deploy. Family and revision (family:revision) or full ARN. Default is none, and use current service's task definition")
	flags.StringVarP(&s.imageWithTag, "image", "i", "", "Name of Docker image to run, ex: repo/image:latest")
	flags.IntVarP(&s.timeout, "timeout", "t", 300, "Timeout seconds. Script monitors ECS Service for new task definition to be running")
	flags.BoolVar(&s.enableRollback, "enable-rollback", false, "Rollback task definition if new version is not running before TIMEOUT")
	flags.BoolVar(&s.skipCheckDeployments, "skip-check-deployments", false, "Skip checking deployments when detect whether deploy completed")

	return cmd
}

func (s *updateService) update(cmd *cobra.Command, args []string) {
	var baseTaskDefinition *string
	if len(s.baseTaskDefinition) > 0 {
		baseTaskDefinition = &s.baseTaskDefinition
	}
	profile, region, verbose := generalConfig()
	if !verbose {
		log.SetLevel(log.ErrorLevel)
	}
	service, err := ecsdeploy.NewService(s.cluster, s.name, s.imageWithTag, baseTaskDefinition, (time.Duration(s.timeout) * time.Second), s.enableRollback, s.skipCheckDeployments, profile, region, verbose)
	if err != nil {
		log.Fatal(err)
	}
	if err := service.Deploy(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deploy success")
}
