package stacks

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"qaz/bucket"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// Deploy - Launch Cloudformation Stack based on config values
func (s *Stack) Deploy() error {

	err := s.DeployTimeParser()
	if err != nil {
		return err
	}

	Log.Debug(fmt.Sprintf("Updated Template:\n%s", s.Template))
	done := make(chan bool)
	svc := cloudformation.New(s.Session, &aws.Config{Credentials: s.creds()})

	createParams := &cloudformation.CreateStackInput{
		StackName:       aws.String(s.Stackname),
		DisableRollback: aws.Bool(s.Rollback),
	}

	if s.Policy != "" {
		if strings.HasPrefix(s.Policy, "http://") || strings.HasPrefix(s.Policy, "https://") {
			createParams.StackPolicyURL = &s.Policy
		} else {
			createParams.StackPolicyBody = &s.Policy
		}
	}

	// NOTE: Add parameters and tags flag here if set
	if len(s.Parameters) > 0 {
		createParams.Parameters = s.Parameters
	}

	if len(s.Tags) > 0 {
		createParams.Tags = s.Tags
	}

	// If IAM is being touched, add Capabilities
	if strings.Contains(s.Template, "AWS::IAM") {
		createParams.Capabilities = []*string{
			aws.String(cloudformation.CapabilityCapabilityIam),
			aws.String(cloudformation.CapabilityCapabilityNamedIam),
		}
	}

	// If bucket - upload to s3
	if s.Bucket != "" {
		exists, err := bucket.Exists(s.Bucket, s.Session)
		if err != nil {
			Log.Warn(fmt.Sprintf("Received Error when checking if [%s] exists: %s", s.Bucket, err.Error()))
		}

		if !exists {
			Log.Info(fmt.Sprintf(("Creating Bucket [%s]"), s.Bucket))
			if err = bucket.Create(s.Bucket, s.Session); err != nil {
				return err
			}
		}
		t := time.Now()
		tStamp := fmt.Sprintf("%d-%d-%d_%d%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
		url, err := bucket.S3write(s.Bucket, fmt.Sprintf("%s_%s.Template", s.Stackname, tStamp), s.Template, s.Session)
		if err != nil {
			return err
		}
		createParams.TemplateURL = &url
	} else {
		createParams.TemplateBody = &s.Template
	}

	Log.Debug(fmt.Sprintln("Calling [CreateStack] with parameters:", createParams))
	if _, err := svc.CreateStack(createParams); err != nil {
		return errors.New(fmt.Sprintln("Deploying failed: ", err.Error()))

	}

	go s.tail("CREATE", done)
	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.Stackname),
	}

	Log.Debug(fmt.Sprintln("Calling [WaitUntilStackCreateComplete] with parameters:", describeStacksInput))
	if err := svc.WaitUntilStackCreateComplete(describeStacksInput); err != nil {
		return err
	}

	Log.Info(fmt.Sprintf("Deployment successful: [%s]", s.Stackname))

	done <- true
	return nil
}
