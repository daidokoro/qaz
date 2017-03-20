package commands

import (
	"strings"
	"testing"
)

// TestStacks - tests stack type and methods
func TestStack(t *testing.T) {

	teststack := stack{
		name: "sqs",
	}

	// Define sources
	testConfigSrc := `s3://daidokoro-dev/qaz/test/config.yml`
	testTemplateSrc := `s3://daidokoro-dev/qaz/test/sqs.yml`

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

	// Get Stack template - test s3Read
	teststack.template, err = genTimeParser(testTemplateSrc)
	if err != nil {
		t.Error(err)
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

	// Test Check/Validate template
	if err := teststack.check(sess); err != nil {
		t.Error(err, "\n", teststack.template)
	}

	// Test State method
	if _, err := teststack.state(sess); err != nil {
		t.Error(err)
	}

	// Test stackExists method
	if ok := teststack.stackExists(sess); !ok {
		t.Error("Expected True for StackExists but got:", ok)
	}

	// Test UpdateStack
	teststack.template = strings.Replace(teststack.template, "MySecret", "Secret", -1)
	if err := teststack.update(sess); err != nil {
		t.Error(err)
	}

	// Test ChangeSets
	teststack.template = strings.Replace(teststack.template, "Secret", "MySecret", -1)
	job.changeName = "gotest"

	for _, c := range []string{"create", "list", "desc", "execute"} {
		if err := teststack.change(sess, c); err != nil {
			t.Error(err)
		}
	}
}

// TestDeploy - test deploy and terminate stack.
func TestDeploy(t *testing.T) {
	teststack := stack{
		name: "vpc",
	}

	// Define sources
	deployTemplateSrc := `https://raw.githubusercontent.com/daidokoro/qaz/master/examples/vpc/templates/vpc.yml`
	deployConfSource := `https://raw.githubusercontent.com/daidokoro/qaz/master/examples/vpc/config.yml`

	// Get Config
	if err := configReader(deployConfSource); err != nil {
		t.Error(err)
	}

	// create session
	sess, err := awsSession()
	if err != nil {
		t.Error(err)
	}

	teststack.setStackName()

	// Get Stack template - test s3Read
	teststack.template, err = genTimeParser(deployTemplateSrc)
	if err != nil {
		t.Error(err)
	}

	// Test Deploy Stack
	if err := teststack.deploy(sess); err != nil {
		t.Error(err)
	}

	// Test Terminate Stack
	if err := teststack.terminate(sess); err != nil {
		t.Error(err)
	}
}
