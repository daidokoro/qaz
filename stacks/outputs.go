package stacks

import (
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

	log.Debug("Calling [DescribeStacks] with parameters: %v", outputParams)
	outputs, err := svc.DescribeStacks(outputParams)
	if err != nil {
		return fmt.Errorf("Unable to reach stack: %v", err)
	}

	// set stack outputs property
	s.Output = outputs

	return nil
}
