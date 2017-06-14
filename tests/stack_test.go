package tests

import (
	"fmt"
	"qaz/logger"
	stks "qaz/stacks"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

var (
	debugmode = true
	colors    = false
	project   = "github-release"

	// define logger
	log = logger.Logger{
		DebugMode: &debugmode,
		Colors:    &colors,
	}
)

var teststack = stks.Stack{
	Name:    "sqs",
	Project: &project,
	Profile: "default",
	Session: session.Must(session.NewSessionWithOptions(
		session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Config: aws.Config{
				Region: aws.String("eu-west-1")},
			Profile: "default",
		})),
}

func TestStackStates(t *testing.T) {

	// define stks logging
	stks.Log = &log

	// stack Name test
	teststack.SetStackName()
	if teststack.Stackname != "github-release-sqs" {
		t.Error(fmt.Errorf(`Setting Stackname failed, expected: [github-release-sqs], found: [%s]`, teststack.Stackname))
	}

	_, err := teststack.State()
	if err != nil {
		t.Error(fmt.Errorf("stack.State test failed: %s", err))
	}

	if err = teststack.Status(); err != nil {
		t.Error(fmt.Errorf("stack.Status test failed: %s", err))
	}
}
