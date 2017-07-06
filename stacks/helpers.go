package stacks

import (
	"fmt"
	"qaz/bucket"
	"time"
)

// cleanup functions in create_failed or delete_failed states
func (s *Stack) cleanup() error {
	Log.Debug(fmt.Sprintf("Running stack cleanup on [%s]", s.Name))
	resp, err := s.State()
	if err != nil {
		return err
	}

	if resp == state.failed {
		if err := s.terminate(); err != nil {
			return err
		}
	}
	return nil
}

// bucket
func resolveBucket(s *Stack) (string, error) {
	exists, err := bucket.Exists(s.Bucket, s.Session)
	if err != nil {
		Log.Warn(fmt.Sprintf("Received Error when checking if [%s] exists: %s", s.Bucket, err.Error()))
	}
	fmt.Println("This is test")
	if !exists {
		Log.Info(fmt.Sprintf(("Creating Bucket [%s]"), s.Bucket))
		if err = bucket.Create(s.Bucket, s.Session); err != nil {
			return "", err
		}
	}
	t := time.Now()
	tStamp := fmt.Sprintf("%d-%d-%d_%d%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
	url, err := bucket.S3write(s.Bucket, fmt.Sprintf("%s_%s.template", s.Stackname, tStamp), s.Template, s.Session)
	if err != nil {
		return "", err
	}
	return url, nil
}
