package stacks

import (
	"fmt"
	"regexp"

	yaml "gopkg.in/yaml.v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/daidokoro/qaz/log"
)

// Check - Validate Cloudformation templates
func (s *Stack) Check() error {
	svc := cloudformation.New(s.Session, &aws.Config{Credentials: s.creds()})

	params := &cloudformation.ValidateTemplateInput{
		TemplateBody: aws.String(s.Template),
	}

	log.Debug("Calling [ValidateTemplate] with parameters:\n%s"+"\n--\n", params)
	resp, err := svc.ValidateTemplate(params)
	if err != nil {
		return err
	}

	log.Info("valid!\n--\n")

	resp.GoString()

	out, err := yaml.Marshal(resp)
	if err != nil {
		return err
	}

	reg, err := regexp.Compile(OutputRegex)
	if err != nil {
		return err
	}

	output := reg.ReplaceAllStringFunc(string(out), func(s string) string {
		return log.ColorString(s, "cyan")
	})

	fmt.Println(output)

	return nil
}
