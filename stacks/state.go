package stacks

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// StackExists - Returns true if stack exists in AWS Account, returns false if err when checking
func (s *Stack) StackExists() bool {
	svc := cloudformation.New(s.Session, &aws.Config{Credentials: s.creds()})

	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.Stackname),
	}

	Log.Debug(fmt.Sprintln("Calling [DescribeStacks] with parameters:", describeStacksInput))
	_, err := svc.DescribeStacks(describeStacksInput)

	if err == nil {
		return true
	}

	return false
}

// State - returns complete/failed/pending state of stack
func (s *Stack) State() (string, error) {
	svc := cloudformation.New(s.Session, &aws.Config{Credentials: s.creds()})

	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.Stackname),
	}

	Log.Debug(fmt.Sprintln("Calling [DescribeStacks] with parameters: ", describeStacksInput))
	status, err := svc.DescribeStacks(describeStacksInput)
	if err != nil {
		if strings.Contains(err.Error(), "not exist") {
			return state.pending, nil
		}
		return "", err
	}

	if strings.Contains(strings.ToLower(status.GoString()), "complete") {
		return state.complete, nil
	} else if strings.Contains(strings.ToLower(status.GoString()), "fail") {
		return state.failed, nil
	}
	return "", nil
}

// ChangeSetStatus - returns the literal change-set status
func (s *Stack) ChangeSetStatus(args ...string) (string, error) {
	svc := cloudformation.New(s.Session, &aws.Config{Credentials: s.creds()})

	params := &cloudformation.DescribeChangeSetInput{
		StackName:     aws.String(s.Stackname),
		ChangeSetName: &args[0],
	}

	Log.Debug(fmt.Sprintln("Calling [DescribeChangeSet] with parameters: ", params))
	status, err := svc.DescribeChangeSet(params)
	if err != nil {
		return "", err
	}

	return *status.Status, nil

}

// StackStatus - return the literal stack status
func (s *Stack) StackStatus(args ...string) (string, error) {
	svc := cloudformation.New(s.Session, &aws.Config{Credentials: s.creds()})
	params := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.Stackname),
	}

	Log.Debug(fmt.Sprintln("Calling [DescribeStacks] with parameters: ", params))
	status, err := svc.DescribeStacks(params)
	if err != nil {
		return "", err
	}

	return *status.Stacks[0].StackStatus, nil
}
