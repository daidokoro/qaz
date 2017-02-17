package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

type function struct {
	name     string
	payload  []byte
	response interface{}
}

func (f *function) Invoke(sess *session.Session) error {
	svc := lambda.New(sess)

	params := &lambda.InvokeInput{
		FunctionName: aws.String(f.name), // Required
		// ClientContext:  aws.String("String"),
		// InvocationType: aws.String("InvocationType"),
		// LogType:        aws.String("LogType"),
		Payload: []byte(f.payload),
		// Qualifier:      aws.String("Qualifier"),
	}

	resp, err := svc.Invoke(params)

	if err != nil {
		return err
	}

	fmt.Println(resp)
	return nil
}

func listLambda(sess *session.Session) error {
	sess, err := session.NewSession()
	if err != nil {
		fmt.Println("failed to create session,", err)
		return nil
	}

	svc := lambda.New(sess)

	params := &lambda.ListFunctionsInput{
		Marker:   aws.String("String"),
		MaxItems: aws.Int64(20),
	}
	resp, err := svc.ListFunctions(params)

	if err != nil {
		return err
	}

	// Pretty-print the response data.
	fmt.Println(resp)
	return nil
}
