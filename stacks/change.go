package stacks

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// Change - Manage Cloudformation Change-Sets
func (s *Stack) Change(req, changename string) error {
	svc := cloudformation.New(s.Session, &aws.Config{Credentials: s.creds()})

	switch req {

	case "create", "transform":
		// Resolve Deploy-Time functions
		err := s.DeployTimeParser()
		if err != nil {
			return err
		}

		params := &cloudformation.CreateChangeSetInput{
			StackName:     aws.String(s.Stackname),
			ChangeSetName: aws.String(changename),
		}

		if req == "transform" {
			params.ChangeSetType = aws.String("CREATE")
		}

		// add tags if set
		if len(s.Tags) > 0 {
			params.Tags = s.Tags
		}

		Log.Debug(fmt.Sprintf("Updated Template:\n%s", s.Template))

		// If bucket - upload to s3
		var url string

		if s.Bucket != "" {
			url, err = resolveBucket(s)
			if err != nil {
				return err
			}
			params.TemplateURL = &url
		} else {
			params.TemplateBody = &s.Template
		}

		// If IAM is bening touched, add Capabilities
		if strings.Contains(s.Template, "AWS::IAM") {
			params.Capabilities = []*string{
				aws.String(cloudformation.CapabilityCapabilityIam),
				aws.String(cloudformation.CapabilityCapabilityNamedIam),
			}
		}

		if _, err = svc.CreateChangeSet(params); err != nil {
			return err
		}

		describeParams := &cloudformation.DescribeChangeSetInput{
			StackName:     aws.String(s.Stackname),
			ChangeSetName: aws.String(changename),
		}

		for {
			// Waiting for PENDING state to change
			resp, err := svc.DescribeChangeSet(describeParams)
			if err != nil {
				return err
			}

			Log.Info(fmt.Sprintf("Creating Change-Set: [%s] - %s - %s", changename, Log.ColorMap(*resp.Status), s.Stackname))

			if *resp.Status == "CREATE_COMPLETE" || *resp.Status == "FAILED" {
				break
			}

			time.Sleep(time.Second * 1)
		}

	case "rm":
		params := &cloudformation.DeleteChangeSetInput{
			ChangeSetName: aws.String(changename),
			StackName:     aws.String(s.Stackname),
		}

		if _, err := svc.DeleteChangeSet(params); err != nil {
			return err
		}

		Log.Info(fmt.Sprintf("Change-Set: [%s] deleted", changename))

	case "list":
		params := &cloudformation.ListChangeSetsInput{
			StackName: aws.String(s.Stackname),
		}

		resp, err := svc.ListChangeSets(params)
		if err != nil {
			return err
		}

		for _, i := range resp.Summaries {
			Log.Info(fmt.Sprintf("%s%s - Change-Set: [%s] - Status: [%s]", Log.ColorString("@", "magenta"), i.CreationTime.Format(time.RFC850), *i.ChangeSetName, *i.ExecutionStatus))
		}

	case "execute":
		done := make(chan bool)
		params := &cloudformation.ExecuteChangeSetInput{
			StackName:     aws.String(s.Stackname),
			ChangeSetName: aws.String(changename),
		}

		if _, err := svc.ExecuteChangeSet(params); err != nil {
			return err
		}

		describeStacksInput := &cloudformation.DescribeStacksInput{
			StackName: aws.String(s.Stackname),
		}

		go s.tail("UPDATE", done)

		Log.Debug(fmt.Sprintln("Calling [WaitUntilStackUpdateComplete] with parameters:", describeStacksInput))
		if err := svc.WaitUntilStackUpdateComplete(describeStacksInput); err != nil {
			return err
		}

		done <- true

	case "desc":
		params := &cloudformation.DescribeChangeSetInput{
			ChangeSetName: aws.String(changename),
			StackName:     aws.String(s.Stackname),
		}

		resp, err := svc.DescribeChangeSet(params)
		if err != nil {
			return err
		}

		o, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", o)
	}

	return nil
}
