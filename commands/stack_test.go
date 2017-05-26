package commands

import (
	"strings"
	"testing"
)

// TestStacks - tests stack type and methods
func TestStack(t *testing.T) {

	teststack := stack{
		name:    "sqs",
		profile: "default",
	}

	// Define sources
	testConfigSrc := `s3://daidokoro-dev/qaz/test/config.yml`
	testTemplateSrc := `s3://daidokoro-dev/qaz/test/sqs.yml`

	// Get Config
	err := configReader(testConfigSrc, run.cfgRaw)
	if err != nil {
		t.Error(err)
	}

	// create session
	teststack.session, err = manager.GetSess(teststack.profile)
	if err != nil {
		t.Error(err)
	}

	// Set stack name
	teststack.setStackName()
	if teststack.stackname != "github-release-sqs" {
		t.Errorf("StackName Failed, Expected: github-release-sqs, Received: %s", teststack.stackname)
	}

	// set template source
	teststack.source = testTemplateSrc
	teststack.template, err = fetchContent(teststack.source)
	if err != nil {
		t.Error(err)
	}

	// Get Stack template - test s3Read
	if err := teststack.genTimeParser(); err != nil {
		t.Error(err)
	}

	// Test Stack status method
	if err := teststack.status(); err != nil {
		t.Error(err)
	}

	// Test Stack output method
	if err := teststack.outputs(); err != nil {
		t.Error(err)
	}

	// Test Stack output length
	if len(teststack.output.Stacks) < 1 {
		t.Errorf("Expected Output Length to be greater than 0: Got: %s", teststack.output.Stacks)
	}

	// Test Check/Validate template
	if err := teststack.check(); err != nil {
		t.Error(err, "\n", teststack.template)
	}

	// Test State method
	if _, err := teststack.state(); err != nil {
		t.Error(err)
	}

	// Test stackExists method
	if ok := teststack.stackExists(); !ok {
		t.Error("Expected True for StackExists but got:", ok)
	}

	// Test UpdateStack
	teststack.template = strings.Replace(teststack.template, "MySecret", "Secret", -1)
	if err := teststack.update(); err != nil {
		t.Error(err)
	}

	// Test ChangeSets
	teststack.template = strings.Replace(teststack.template, "Secret", "MySecret", -1)
	run.changeName = "gotest"

	for _, c := range []string{"create", "list", "desc", "execute"} {
		if err := teststack.change(c); err != nil {
			t.Error(err)
		}
	}

	return
}

// TestDeploy - test deploy and terminate stack.
func TestDeploy(t *testing.T) {
	run.debug = true
	teststack := stack{
		name:    "vpc",
		profile: "default",
	}

	// Define sources
	deployTemplateSrc := `https://raw.githubusercontent.com/daidokoro/qaz/master/examples/vpc/templates/vpc.yml`
	deployConfSource := `https://raw.githubusercontent.com/daidokoro/qaz/master/examples/vpc/config.yml`

	// Get Config
	err := configReader(deployConfSource, run.cfgRaw)
	if err != nil {
		t.Error(err)
	}

	// create session
	teststack.session, err = manager.GetSess(teststack.profile)
	if err != nil {
		t.Error(err)
	}

	teststack.setStackName()

	// Set source
	teststack.source = deployTemplateSrc
	resp, err := fetchContent(teststack.source)
	if err != nil {
		t.Error(err)
	}

	teststack.template = resp

	// Get Stack template - test s3Read
	if err = teststack.genTimeParser(); err != nil {
		t.Error(err)
	}

	// Test Deploy Stack
	if err := teststack.deploy(); err != nil {
		t.Error(err)
	}

	// Test Set Stack Policy
	teststack.policy = stacks[teststack.name].policy
	if err := teststack.stackPolicy(); err != nil {
		t.Errorf("%s - [%s]", err, teststack.policy)
	}

	// Test Terminate Stack
	if err := teststack.terminate(); err != nil {
		t.Error(err)
	}

	return
}
