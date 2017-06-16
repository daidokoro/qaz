package stacks

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// StackPolicy - Stack Cloudformation Stack policy
func (s *Stack) StackPolicy() error {

	if s.Policy == "" {
		return fmt.Errorf("Empty Stack Policy value detected...")
	}

	svc := cloudformation.New(s.Session, &aws.Config{Credentials: s.creds()})

	params := &cloudformation.SetStackPolicyInput{
		StackName: &s.Stackname,
	}

	// Check if source is a URL
	if strings.HasPrefix(s.Policy, `http://`) || strings.HasPrefix(s.Policy, `https://`) {
		params.StackPolicyURL = &s.Policy
	} else {
		params.StackPolicyBody = &s.Policy
	}

	Log.Debug(fmt.Sprintln("Calling SetStackPolicy with params: ", params))
	resp, err := svc.SetStackPolicy(params)
	if err != nil {
		return err
	}

	Log.Info(fmt.Sprintf("Stack Policy applied: [%s] - %s", s.Stackname, resp.GoString()))

	return nil
}
