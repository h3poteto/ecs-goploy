package deploy

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	eventsiface "github.com/aws/aws-sdk-go/service/cloudwatchevents/cloudwatcheventsiface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ScheduledTask has target task definition information and client of aws-sdk-go.
type ScheduledTask struct {
	awsCloudWatchEvents eventsiface.CloudWatchEventsAPI

	// TaskDefinition struct to call aws API.
	TaskDefinition *TaskDefinition

	verbose bool
}

// NewScheduledTask returns a nwe ScheduledTask struct, and initialize aws cloudwatchevents API client.
func NewScheduledTask(profile, region string, verbose bool) *ScheduledTask {
	awsCloudWatchEvents := events.New(session.New(), newConfig(profile, region))
	taskDefinition := NewTaskDefinition(profile, region, verbose)
	if !verbose {
		log.SetLevel(log.ErrorLevel)
	}
	return &ScheduledTask{
		awsCloudWatchEvents,
		taskDefinition,
		verbose,
	}
}

// ListsEventTargets list up event targets based on rule name.
func (s *ScheduledTask) ListsEventTargets(ruleName *string) ([]*events.Target, error) {
	params := &events.ListTargetsByRuleInput{
		Rule: ruleName,
	}
	resp, err := s.awsCloudWatchEvents.ListTargetsByRule(params)
	if err != nil {
		return nil, err
	}
	return resp.Targets, nil
}

// DescribeRule finds an event rule.
func (s *ScheduledTask) DescribeRule(name string) (*events.DescribeRuleOutput, error) {
	params := &events.DescribeRuleInput{
		Name: aws.String(name),
	}
	resp, err := s.awsCloudWatchEvents.DescribeRule(params)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// update updates an event target.
func (s *ScheduledTask) update(taskCount int64, taskDefinition *ecs.TaskDefinition, baseTarget *events.Target, ruleName *string) error {
	ecsParameter := &events.EcsParameters{
		TaskCount:         aws.Int64(taskCount),
		TaskDefinitionArn: taskDefinition.TaskDefinitionArn,
	}
	target := baseTarget.SetEcsParameters(ecsParameter)
	params := &events.PutTargetsInput{
		Rule: ruleName,
		Targets: []*events.Target{
			target,
		},
	}
	resp, err := s.awsCloudWatchEvents.PutTargets(params)
	if err != nil {
		return err
	}
	if *resp.FailedEntryCount > 0 {
		for _, e := range resp.FailedEntries {
			log.Errorf("Failed to update the entry: %+v", *e)
		}
		return errors.New("Failed to update entries")
	}
	return nil
}

// UpdateTargets updates all event targets related the rule.
func (s *ScheduledTask) UpdateTargets(taskCount int64, taskDefinition *ecs.TaskDefinition, name string) error {
	rule, err := s.DescribeRule(name)
	if err != nil {
		return err
	}
	targets, err := s.ListsEventTargets(rule.Name)
	if err != nil {
		return err
	}
	for _, target := range targets {
		err := s.update(taskCount, taskDefinition, target, rule.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
