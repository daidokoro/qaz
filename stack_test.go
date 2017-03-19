package main

import "testing"

var teststack = stack{
	name: "sqs",
}

var testConfigSrc = `s3://daidokoro-dev/qaz/test/config.yml`

// TestStacks - tests stack type and methods
func TestStack(t *testing.T) {
	// Get Config
	if err := configReader(testConfigSrc); err != nil {
		t.Error(err)
	}

	// create session
	sess, err := awsSession()
	if err != nil {
		t.Error(err)
	}

	// Set stack name
	teststack.setStackName()
	if teststack.stackname != "github-release-sqs" {
		t.Errorf("StackName Failed, Expected: github-release-sqs, Received: %s", teststack.stackname)
	}

	// Test Stack status method
	if err := teststack.status(sess); err != nil {
		t.Error(err)
	}

	// Test Stack output method
	if err := teststack.outputs(sess); err != nil {
		t.Error(err)
	}

	// Test Stack output length
	if len(teststack.output.Stacks) < 1 {
		t.Errorf("Expected Output Length to be greater than 0: Got: %s", teststack.output.Stacks)
	}

}
