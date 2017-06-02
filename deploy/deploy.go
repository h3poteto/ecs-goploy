package deploy

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// Deploy have target ecs information
type Deploy struct {
	awsECS  *ecs.ECS
	cluster string
	name    string
	image   string
}

// NewDeploy return a new Deploy struct, and initialize aws ecs api client
func NewDeploy(cluster, name, image, profile, region string) *Deploy {
	awsECS := ecs.New(session.New(), newConfig(profile, region))
	return &Deploy{
		awsECS,
		cluster,
		name,
		image,
	}
}

// Deploy run deploy commands
func (d *Deploy) Deploy() {
	if err := d.TaskDefinition(); err != nil {
		log.Fatalln(err)
	}
}
