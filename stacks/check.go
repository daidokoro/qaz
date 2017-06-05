package stacks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// Check - Validate Cloudformation templates
func (s *Stack) Check() error {
	svc := cloudformation.New(s.Session, &aws.Config{Credentials: s.creds()})

	params := &cloudformation.ValidateTemplateInput{
		TemplateBody: aws.String(s.Template),
	}

	Log.Debug(fmt.Sprintf("Calling [ValidateTemplate] with parameters:\n%s"+"\n--\n", params))
	resp, err := svc.ValidateTemplate(params)
	if err != nil {
		return err
	}

	fmt.Printf(
		"%s\n\n%s"+"\n",
		Log.ColorString("Valid!", "green"),
		resp.GoString(),
	)

	return nil
}
