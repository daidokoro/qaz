package troposphere

import (
	"testing"

	"github.com/daidokoro/qaz/logger"
)

const testcode = `
from troposphere import Output, Ref, Template
from troposphere.s3 import Bucket, PublicRead

t = Template()

s3bucket = t.add_resource(Bucket("S3Bucket", AccessControl=PublicRead,))

t.add_output(Output(
    "BucketName",
    Value=Ref(s3bucket),
    Description="Name of S3 bucket to hold website content"
))

print(t.to_json())`

func TestBuild(t *testing.T) {
	Logger(logger.New(false, false))
	if err := BuildImage(); err != nil {
		t.Error(err)
	}
}

func TestExecute(t *testing.T) {
	Logger(logger.New(false, true))
	s, err := Execute(testcode)
	if err != nil {
		t.Error(err)
	}

	t.Log(s)
}
