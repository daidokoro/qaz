package stacks

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func (s *Stack) terminate() error {

	if !s.StackExists() {
		Log.Info(fmt.Sprintf("%s: does not exist...", s.Name))
		return nil
	}

	done := make(chan bool)
	svc := cloudformation.New(s.Session, &aws.Config{Credentials: s.creds()})

	params := &cloudformation.DeleteStackInput{
		StackName: aws.String(s.Stackname),
	}

	// create wait handler for tail
	var tailinput = TailServiceInput{
		printed: make(map[string]interface{}),
		stk:     *s,
		command: "DELETE",
	}

	go tailWait(done, &tailinput)

	Log.Debug(fmt.Sprintln("Calling [DeleteStack] with parameters:", params))
	if _, err := svc.DeleteStack(params); err != nil {
		done <- true
		return errors.New(fmt.Sprintln("Deleting failed: ", err))
	}

	if err := svc.WaitUntilStackDeleteComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(s.Stackname),
	}); err != nil {
		return err
	}

	done <- true
	Log.Info(fmt.Sprintf("Deletion successful: [%s]", s.Stackname))

	return nil
}
