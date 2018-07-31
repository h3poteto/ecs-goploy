package deploy

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	eventsiface "github.com/aws/aws-sdk-go/service/cloudwatchevents/cloudwatcheventsiface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/pkg/errors"
)

type ScheduledTask struct {
	awsCloudWatchEvents eventsiface.CloudWatchEventsAPI

	// TaskDefinition struct to call aws API.
	TaskDefinition *TaskDefinition
}

func NewScheduledTask(profile, region string) *ScheduledTask {
	awsCloudWatchEvents := events.New(session.New(), newConfig(profile, region))
	taskDefinition := NewTaskDefinition(profile, region)
	return &ScheduledTask{
		awsCloudWatchEvents,
		taskDefinition,
	}
}

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
			log.Printf("[ERROR] Failed to update the entry: %+v\n", *e)
		}
		return errors.New("Failed to update entries")
	}
	return nil
}

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
