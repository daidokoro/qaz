package bucket

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"qaz/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// -- Contains all things S3

// define logger
var log *logger.Logger

// S3Read - Reads the content of a given s3 url endpoint and returns the content string.
func S3Read(url string, sess *session.Session) (string, error) {
	svc := s3.New(sess)

	// Parse s3 url
	bucket := strings.Split(strings.Replace(strings.ToLower(url), `s3://`, "", -1), `/`)[0]
	key := strings.Replace(strings.ToLower(url), fmt.Sprintf("s3://%s/", bucket), "", -1)

	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	log.Debug(fmt.Sprintln("Calling S3 [GetObject] with parameters:", params))
	resp, err := svc.GetObject(params)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)

	log.Debug("Reading from S3 Response Body")
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

	log.Debug(fmt.Sprintln("Calling S3 [PutObject] with parameters:", params))
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

	log.Debug(fmt.Sprintln("Calling S3 [CreateBucket] with parameters:", params))
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

	log.Debug(fmt.Sprintln("Calling S3 [HeadBucket] with parameters:", params))
	_, err := svc.HeadBucket(params)
	if err != nil {
		return false, err
	}

	return true, nil

}
