package bucket

import (
	"bytes"
	"net/url"
	"strings"
	"time"

	"github.com/daidokoro/qaz/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// -- Contains all things S3

// Log define logger
var Log *logger.Logger

// S3Read - Reads the content of a given s3 url endpoint and returns the content string.
func S3Read(URL string, sess *session.Session) (string, error) {
	svc := s3.New(sess)

	src, err := url.Parse(URL)
	if err != nil {
		return "", err
	}

	// Parse s3 url
	bucket := src.Scheme
	key := strings.TrimPrefix(src.Path, "/")

	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	Log.Debug("calling S3 [GetObject] with parameters: %s", params)
	resp, err := svc.GetObject(params)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)

	Log.Debug("Reading from S3 Response Body")
	buf.ReadFrom(resp.Body)
	return buf.String(), nil
}

// S3write - Writes a file to s3 and returns the presigned url
func S3write(bucket string, key string, body string, sess *session.Session) (string, error) {
	svc := s3.New(sess)
	params := &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   bytes.NewReader([]byte(body)),
		Metadata: map[string]*string{
			"created_by": aws.String("qaz"),
		},
	}

	Log.Debug("calling S3 [PutObject] with parameters: %s", params)
	_, err := svc.PutObject(params)
	if err != nil {
		return "", err
	}

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	url, err := req.Presign(10 * time.Minute)
	if err != nil {
		return "", err
	}

	return url, nil
}

// Create - create s3 bucket
func Create(bucket string, sess *session.Session) error {
	svc := s3.New(sess)

	params := &s3.CreateBucketInput{
		Bucket: &bucket,
	}

	Log.Debug("calling S3 [CreateBucket] with parameters: %s", params)
	_, err := svc.CreateBucket(params)
	if err != nil {
		return err
	}

	if err := svc.WaitUntilBucketExists(&s3.HeadBucketInput{Bucket: aws.String(bucket)}); err != nil {
		return err
	}

	return nil
}

// Exists - checks if bucket exists - if err, then its assumed that the bucket does not exist.
func Exists(bucket string, sess *session.Session) (bool, error) {
	svc := s3.New(sess)
	params := &s3.HeadBucketInput{
		Bucket: &bucket,
	}

	Log.Debug("calling S3 [HeadBucket] with parameters: %s", params)
	_, err := svc.HeadBucket(params)
	if err != nil {
		return false, err
	}

	return true, nil

}
