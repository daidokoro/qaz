package commands

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

type awsLambda struct {
	name     string
	payload  []byte
	response string
}

func (a *awsLambda) Invoke(sess *session.Session) error {
	svc := lambda.New(sess)

	params := &lambda.InvokeInput{
		FunctionName: aws.String(a.name),
	}

	if a.payload != nil {
		params.Payload = a.payload
	}

	Log(fmt.Sprintln("Calling [Invoke] with parameters:", params), level.debug)
	resp, err := svc.Invoke(params)

	if err != nil {
		return err
	}

	if resp.FunctionError != nil {
		return fmt.Errorf(*resp.FunctionError)
	}

	a.response = string(resp.Payload)

	Log(fmt.Sprintln("Lambda response:", a.response), level.debug)
	return nil
}
