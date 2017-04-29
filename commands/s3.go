package commands

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// -- Contains all things S3

// S3Read - Reads the content of a given s3 url endpoint and returns the content string.
func S3Read(url string) (string, error) {

	sess, err := manager.GetSess(job.profile)
	if err != nil {
		return "", err
	}

	svc := s3.New(sess)

	// Parse s3 url
	bucket := strings.Split(strings.Replace(strings.ToLower(url), `s3://`, "", -1), `/`)[0]
	key := strings.Replace(strings.ToLower(url), fmt.Sprintf("s3://%s/", bucket), "", -1)

	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	Log(fmt.Sprintln("Calling S3 [GetObject] with parameters:", params), level.debug)
	resp, err := svc.GetObject(params)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)

	Log("Reading from S3 Response Body", level.debug)
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

	Log(fmt.Sprintln("Calling S3 [PutObject] with parameters:", params), level.debug)
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

// CreateBucket - create s3 bucket
func CreateBucket(bucket string, sess *session.Session) error {
	svc := s3.New(sess)

	params := &s3.CreateBucketInput{
		Bucket: &bucket,
	}

	Log(fmt.Sprintln("Calling S3 [CreateBucket] with parameters:", params), level.debug)
	_, err := svc.CreateBucket(params)
	if err != nil {
		return err
	}

	if err := svc.WaitUntilBucketExists(&s3.HeadBucketInput{Bucket: aws.String(bucket)}); err != nil {
		return err
	}

	return nil
}

// BucketExists - checks if bucket exists - if err, then its assumed that the bucket does not exist.
func BucketExists(bucket string, sess *session.Session) (bool, error) {
	svc := s3.New(sess)
	params := &s3.HeadBucketInput{
		Bucket: &bucket,
	}

	Log(fmt.Sprintln("Calling S3 [HeadBucket] with parameters:", params), level.debug)
	_, err := svc.HeadBucket(params)
	if err != nil {
		return false, err
	}

	return true, nil

}
