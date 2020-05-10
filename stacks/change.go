package stacks

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/daidokoro/qaz/utils"

	yaml "gopkg.in/yaml.v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

const (
	transform        = "transform"
	create           = "create"
	rm               = "rm"
	list             = "list"
	execute          = "execute"
	desc             = "desc"
	serverless       = "serverless"
	iamCapable       = "AWS::IAM"
	transformCapable = "AWS::Serverless"
)

// Change - Manage Cloudformation Change-Sets
func (s *Stack) Change(req, changename string) error {
	svc := cloudformation.New(s.Session, &aws.Config{Credentials: s.creds()})

	switch req {

	case create, transform:
		// Resolve Deploy-Time functions
		err := s.DeployTimeParser()
		if err != nil {
			return err
		}

		params := &cloudformation.CreateChangeSetInput{
			StackName:     aws.String(s.Stackname),
			ChangeSetName: aws.String(changename),
		}

		if req == transform {
			params.ChangeSetType = aws.String("CREATE")
		}

		// add tags if set
		if len(s.Tags) > 0 {
			params.Tags = s.Tags
		}

		Log.Debug("updated template:\n%s", s.Template)

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

		// NOTE: Add parameters and tags flag here if set
		if len(s.Parameters) > 0 {
			params.Parameters = s.Parameters
		}

		if len(s.Tags) > 0 {
			params.Tags = s.Tags
		}

		// If IAM is bening touched, add Capabilities
		if strings.Contains(s.Template, iamCapable) || strings.Contains(s.Template, transformCapable) {
			params.Capabilities = []*string{
				aws.String(cloudformation.CapabilityCapabilityIam),
				aws.String(cloudformation.CapabilityCapabilityNamedIam),
			}
		}

		Log.Debug("calling [CreateChangeSet] with parameters: %s", params)
		if _, err = svc.CreateChangeSet(params); err != nil {
			return err
		}

		describeParams := &cloudformation.DescribeChangeSetInput{
			StackName:     aws.String(s.Stackname),
			ChangeSetName: aws.String(changename),
		}

		Log.Info("creating change-set: [%s] - %s", changename, s.Stackname)
		if err = Wait(s.ChangeSetStatus, changename); err != nil {
			return err
		}

		resp, err := svc.DescribeChangeSet(describeParams)
		if err != nil {
			return err
		}

		// print reason for FAILED status (e.g. if no changes were detected)
		if *resp.Status == "FAILED" {
			Log.Info(*resp.StatusReason)
		}

		Log.Info("created change-set: [%s] - %s - %s", changename, Log.ColorMap(*resp.Status), s.Stackname)
		return nil

	case rm:
		params := &cloudformation.DeleteChangeSetInput{
			ChangeSetName: aws.String(changename),
			StackName:     aws.String(s.Stackname),
		}

		if _, err := svc.DeleteChangeSet(params); err != nil {
			return err
		}

		Log.Info("change-Set: [%s] deleted", changename)

	case list:
		params := &cloudformation.ListChangeSetsInput{
			StackName: aws.String(s.Stackname),
		}

		resp, err := svc.ListChangeSets(params)
		if err != nil {
			return err
		}

		for _, i := range resp.Summaries {
			Log.Info("%s%s - Change-Set: [%s] - Status: [%s]", Log.ColorString("@", "magenta"), i.CreationTime.Format(time.RFC850), *i.ChangeSetName, *i.ExecutionStatus)
		}

	case execute, serverless:
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

		if req != serverless {
			go s.tail("UPDATE", done)

			Log.Debug("calling [WaitUntilStackUpdateComplete] with parameters:", describeStacksInput)
			if err := svc.WaitUntilStackUpdateComplete(describeStacksInput); err != nil {
				return err
			}
			done <- true
		}

		Log.Info("change-set executed successfully")

	case desc:
		params := &cloudformation.DescribeChangeSetInput{
			ChangeSetName: aws.String(changename),
			StackName:     aws.String(s.Stackname),
		}

		resp, err := svc.DescribeChangeSet(params)
		if err != nil {
			return err
		}

		o, err := yaml.Marshal(resp.Changes)
		if err != nil {
			return err
		}

		reg, err := regexp.Compile(OutputRegex)
		utils.HandleError(err)

		out := reg.ReplaceAllStringFunc(string(o), func(s string) string {
			return Log.ColorString(s, "cyan")
		})

		fmt.Printf(out)
	}

	return nil
}
