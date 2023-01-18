package stacks

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/daidokoro/qaz/log"
	"github.com/fatih/color"
)

// TODO: this function's pretty bad, waaaay too long, need to break it apart

// Deploy - Launch Cloudformation Stack based on config values
func (s *Stack) Deploy() error {

	// if serverless deploy
	if strings.Contains(s.Template, "AWS::Serverless") {
		return s.DeploySAM()
	}

	err := s.DeployTimeParser()
	if err != nil {
		return err
	}

	log.Debug("Updated Template:\n%s", s.Template)
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

	// add timeout if set
	if s.Timeout > 0 {
		createParams.TimeoutInMinutes = aws.Int64(s.Timeout)
	}

	// add notification arns
	if len(s.NotificationARNs) > 0 {
		createParams.NotificationARNs = aws.StringSlice(s.NotificationARNs)
	}

	// If IAM is being touched, add Capabilities
	if strings.Contains(s.Template, "AWS::IAM") {
		createParams.Capabilities = []*string{
			aws.String(cloudformation.CapabilityCapabilityIam),
			aws.String(cloudformation.CapabilityCapabilityNamedIam),
		}
	}

	createParams.Capabilities = append(createParams.Capabilities, aws.String(cloudformation.CapabilityCapabilityAutoExpand))

	// If bucket - upload to s3
	if s.Bucket != "" {
		var url string
		url, err = resolveBucket(s)
		if err != nil {
			return err
		}
		createParams.TemplateURL = &url
	} else {
		createParams.TemplateBody = &s.Template
	}

	log.Debug("Calling [CreateStack] with parameters: %s", createParams)
	if _, err = svc.CreateStack(createParams); err != nil {
		return errors.New(fmt.Sprintln("Deploying failed: ", err.Error()))

	}

	// go s.tail("CREATE", done)
	var tailinput = TailServiceInput{
		printed: make(map[string]interface{}),
		stk:     *s,
		command: "CREATE",
	}

	go tailWait(done, &tailinput)

	err = svc.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(s.Stackname),
	})

	if err != nil {
		return err
	}

	done <- true
	log.Info(
		"deployment completed: %s",
		color.New(color.FgWhite).Add(color.Bold).SprintFunc()(fmt.Sprintf("[%s]", s.Stackname)),
	)

	return nil
}
