package stacks

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// Protect - enables stack termination-protection
func (s *Stack) Protect(enable *bool) error {
	svc := cloudformation.New(s.Session, &aws.Config{Credentials: s.creds()})

	params := &cloudformation.UpdateTerminationProtectionInput{
		EnableTerminationProtection: aws.Bool(!*enable),
		StackName:                   &s.Stackname,
	}

	log.Debug("Calling [UpdateTerminationProtection] with parameters:\n%s"+"\n--\n", params)
	if _, err := svc.UpdateTerminationProtection(params); err != nil {
		return err
	}

	return nil
}
