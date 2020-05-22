package testing

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/daidokoro/qaz/bucket"
	"github.com/stretchr/testify/assert"
)

const (
	// a dummy object existing in S3
	s3uri = "s3://daidokoro-dev/qaz_testing.txt"
	body  = "3347e9fbc394bd6110c8a57da41c4abf"
)

func TestS3Write(t *testing.T) {
	s, err := bucket.S3write(
		"daidokoro-dev",
		"qaz_testing.txt",
		strings.NewReader(body),
		sess,
	)

	assert.NoError(t, err)
	_, err = url.Parse(s)
	// check return uri is valid
	assert.NoError(t, err)
}

func TestS3Read(t *testing.T) {
	s, err := bucket.S3Read(s3uri, sess)
	assert.NoError(t, err)
	assert.Equal(t, "3347e9fbc394bd6110c8a57da41c4abf", s)
}

func TestS3Create(t *testing.T) {
	ti := time.Now()
	b := fmt.Sprintf("qaz-test-bucket-%s", ti.Format("200601021504"))
	assert.NoError(t, bucket.Create(b, sess))

	// test exist
	r, err := bucket.Exists(b, sess)
	assert.NoError(t, err)
	assert.Equal(t, true, r)

	// remove bucket when done
	svc := s3.New(sess)
	_, err = svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: &b,
	})
	assert.NoError(t, err)
}
