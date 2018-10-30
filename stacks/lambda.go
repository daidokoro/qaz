package stacks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

type awslambda struct {
	name     string
	payload  []byte
	response string
}

func (a *awslambda) Invoke(sess *session.Session) error {
	svc := lambda.New(sess)

	params := &lambda.InvokeInput{
		FunctionName: aws.String(a.name),
	}

	if a.payload != nil {
		params.Payload = a.payload
	}

	log.Debug("Calling [Invoke] with parameters: %s", params)
	resp, err := svc.Invoke(params)

	if err != nil {
		return err
	}

	if resp.FunctionError != nil {
		return fmt.Errorf(*resp.FunctionError)
	}

	a.response = string(resp.Payload)

	log.Debug("Lambda response: %s", a.response)
	return nil
}
