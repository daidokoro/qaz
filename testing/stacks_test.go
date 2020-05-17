package testing

import (
	"testing"

	"github.com/daidokoro/qaz/commands"
	"github.com/stretchr/testify/assert"
)

var testConfigSource = "https://raw.githubusercontent.com/daidokoro/qaz/master/examples/vpc/config.yml"

func TestGenerate(t *testing.T) {
	expected := `AWSTemplateFormatVersion: '2010-09-09'

Description: |
  This is an example VPC deployed via Qaz!

Resources:
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.10.0.0/16

Outputs:
  vpcid:
    Description: VPC ID
    Value:  !Ref VPC
    Export:
      Name: !Sub "${AWS::StackName}-vpcid"
`
	stks, err := commands.Configure(testConfigSource, "")
	assert.NoError(t, err)

	// deploying stack
	err = stks.MustGet("vpc").GenTimeParser()
	assert.NoError(t, err)
	assert.Equal(t, expected, stks.MustGet("vpc").Template)
}

// func TestDeploy(t *testing.T) {
// 	stks, err := commands.Configure(testConfigSource, "")
// 	assert.NoError(t, err)

// 	stks.MustGet("vpc").GenTimeParser()

// 	// deploying stack
// 	err = stks.MustGet("vpc").Deploy()
// 	assert.NoError(t, err)
// }

// func TestExists(t *testing.T) {
// 	stks, err := commands.Configure(testConfigSource, "")
// 	assert.NoError(t, err)

// 	// deploying stack
// 	r := stks.MustGet("vpc").StackExists()
// 	assert.Equal(t, true, r)
// }

// func TestTerminate(t *testing.T) {
// 	stks, err := commands.Configure(testConfigSource, "")
// 	assert.NoError(t, err)

// 	// deploying stack
// 	err = stks.MustGet("vpc").Terminate()
// 	assert.NoError(t, err)
// }
