package stacks

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// Outputs - Get Stack outputs
func (s *Stack) Outputs() error {

	svc := cloudformation.New(s.Session, &aws.Config{Credentials: s.creds()})
	outputParams := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.Stackname),
	}

	Log.Debug(fmt.Sprintln("Calling [DescribeStacks] with parameters:", outputParams))
	outputs, err := svc.DescribeStacks(outputParams)
	if err != nil {
		return errors.New(fmt.Sprintln("Unable to reach stack", err.Error()))
	}

	// set stack outputs property
	s.Output = outputs

	return nil
}
