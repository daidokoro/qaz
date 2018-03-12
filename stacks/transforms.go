package stacks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// support for SAM - Serverless Arch Model Cloudformation templates

// DeploySAM deploys SAMs Cloudformation templates
func (s *Stack) DeploySAM() error {
	changename := fmt.Sprintf("%s-change-set", s.Stackname)
	Log.Info(
		"%s [SAM] deploy detected via [%s]: deploying serverless template via change-set",
		Log.ColorString("serverless", "cyan"),
		s.Stackname,
	)

	if err := s.Change(transform, changename); err != nil {
		return err
	}

	if err := s.Change(serverless, changename); err != nil {
		return err
	}

	done := make(chan bool)
	svc := cloudformation.New(s.Session, &aws.Config{Credentials: s.creds()})
	go s.tail("CREATE", done)
	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.Stackname),
	}

	Log.Debug("Calling [WaitUntilStackCreateComplete] with parameters: %s", describeStacksInput)
	if err := svc.WaitUntilStackCreateComplete(describeStacksInput); err != nil {
		return err
	}

	done <- true
	Log.Info(
		"%s [SAM] - deploy completed - %s",
		Log.ColorString("serverless", "cyan"),
		s.Stackname,
	)
	return nil
}
