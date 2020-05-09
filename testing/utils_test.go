package testing

import (
	"testing"

	"github.com/daidokoro/qaz/utils"
	"github.com/stretchr/testify/assert"
)

func TestConfigTemplate(t *testing.T) {
	expected := []byte(`
# AWS Region
region: eu-west-1

# Project Name
project: test-project

# Global Stack Variables
global:

# Stacks
stacks:

`)

	b := utils.ConfigTemplate("test-project", "eu-west-1")
	assert.Equal(t, expected, b)
}

func TestAllinSlice(t *testing.T) {
	slice := []string{"a", "a", "a"}
	assert.Equal(t, true, utils.All(slice, "a"))
	assert.Equal(t, false, utils.All(slice, "b"))
}

func TestStringIn(t *testing.T) {
	slice := []string{"a", "b", "c"}
	assert.Equal(t, true, utils.StringIn("a", slice))
	assert.Equal(t, false, utils.StringIn("q", slice))
}

func TestGet(t *testing.T) {
	expected := "Description: This is an example VPC deployed via Qaz!\n\nResources:\n  VPC:\n    Type: AWS::EC2::VPC\n    Properties:\n      CidrBlock: {{ .vpc.cidr }}\n"
	v, err := utils.Get("https://raw.githubusercontent.com/cfn-deployable/simplevpc/master/vpc.yaml")
	assert.NoError(t, err)
	assert.Equal(t, expected, v)
}

func TestGetSource(t *testing.T) {
	x, y, err := utils.GetSource("Stackname::http://someurl")
	assert.NoError(t, err)
	assert.Equal(t, x, "Stackname")
	assert.Equal(t, y, "http://someurl")
}

func TestIsJSON(t *testing.T) {
	j := `{"name": "some json"}`
	assert.Equal(t, true, utils.IsJSON(j))
}

func TestIsHCL(t *testing.T) {
	h := `service { name = "some hcl" }`
	assert.Equal(t, true, utils.IsHCL(h))
}
