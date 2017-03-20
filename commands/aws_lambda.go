package commands

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

type function struct {
	name     string
	payload  []byte
	response string
}

func (f *function) Invoke(sess *session.Session) error {
	svc := lambda.New(sess)

	params := &lambda.InvokeInput{
		FunctionName: aws.String(f.name),
	}

	if f.payload != nil {
		params.Payload = f.payload
	}

	Log(fmt.Sprintln("Calling [Invoke] with parameters:", params), level.debug)
	resp, err := svc.Invoke(params)

	if err != nil {
		return err
	}

	if resp.FunctionError != nil {
		return fmt.Errorf(*resp.FunctionError)
	}

	f.response = string(resp.Payload)

	Log(fmt.Sprintln("Lambda response:", f.response), level.debug)
	return nil
}
