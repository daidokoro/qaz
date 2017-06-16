package stacks

import (
	"errors"
	"fmt"
	"time"

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

	Log.Debug(fmt.Sprintln("Calling [DeleteStack] with parameters:", params))
	_, err := svc.DeleteStack(params)

	go s.tail("DELETE", done)

	if err != nil {
		done <- true
		return errors.New(fmt.Sprintln("Deleting failed: ", err))
	}

	// describeStacksInput := &cloudformation.DescribeStacksInput{
	// 	StackName: aws.String(s.Stackname),
	// }
	//
	// Log(fmt.Sprintln("Calling [WaitUntilStackDeleteComplete] with parameters:", describeStacksInput), level.debug)
	//
	// if err := svc.WaitUntilStackDeleteComplete(describeStacksInput); err != nil {
	// 	return err
	// }

	// NOTE: The [WaitUntilStackDeleteComplete] api call suddenly stopped playing nice.
	// Implemented this crude loop as a patch fix for now
	for {
		if !s.StackExists() {
			done <- true
			break
		}

		time.Sleep(time.Second * 1)
	}

	Log.Info(fmt.Sprintf("Deletion successful: [%s]", s.Stackname))

	return nil
}
