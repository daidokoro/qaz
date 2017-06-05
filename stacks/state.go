package stacks

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// StackExists - Returns true if stack exists in AWS Account
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

func (s *Stack) state() (string, error) {
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
