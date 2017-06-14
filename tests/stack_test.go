package tests

import (
	"fmt"
	"qaz/logger"
	stks "qaz/stacks"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// NOTE: I need to write more tests. Following the restructuring of the project, all previous tests are no long applicable. The reason for the tests
// package is that most packages within qaz directly or explicitly rely on other packages within the project, making it difficult to write isolated
// tests. Using this package approach I can import dependencies and run tests without conflicts. The downside is that there will be no way
// to properly track coverage.

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
