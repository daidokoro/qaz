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

// Update - Update Cloudformation Stack
func (s *Stack) Update() error {

	err := s.DeployTimeParser()
	if err != nil {
		return err
	}

	done := make(chan bool)
	svc := cloudformation.New(s.Session, &aws.Config{Credentials: s.creds()})
	updateParams := &cloudformation.UpdateStackInput{
		StackName:    aws.String(s.Stackname),
		TemplateBody: aws.String(s.Template),
	}

	// NOTE: Add parameters and tags flag here if set
	if len(s.Parameters) > 0 {
		updateParams.Parameters = s.Parameters
	}

	if len(s.Tags) > 0 {
		updateParams.Tags = s.Tags
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
		updateParams.TemplateURL = &url
	} else {
		updateParams.TemplateBody = &s.Template
	}

	// NOTE: Add parameters flag here if params set
	if len(s.Parameters) > 0 {
		updateParams.Parameters = s.Parameters
	}

	// If IAM is being touched, add Capabilities
	if strings.Contains(s.Template, "AWS::IAM") {
		updateParams.Capabilities = []*string{
			aws.String(cloudformation.CapabilityCapabilityIam),
			aws.String(cloudformation.CapabilityCapabilityNamedIam),
		}
	}

	if s.StackExists() {
		Log.Info("Stack exists, updating...")

		Log.Debug(fmt.Sprintln("Calling [UpdateStack] with parameters:", updateParams))
		_, err := svc.UpdateStack(updateParams)

		if err != nil {
			return errors.New(fmt.Sprintln("Update failed: ", err))
		}

		go s.tail("UPDATE", done)

		describeStacksInput := &cloudformation.DescribeStacksInput{
			StackName: aws.String(s.Stackname),
		}
		Log.Debug(fmt.Sprintln("Calling [WaitUntilStackUpdateComplete] with parameters:", describeStacksInput))
		if err := svc.WaitUntilStackUpdateComplete(describeStacksInput); err != nil {
			return err
		}

		Log.Info(fmt.Sprintf("Stack update successful: [%s]", s.Stackname))

	}
	done <- true
	return nil
}
