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

// Wait - wait Until status is complete
func Wait(getStatus func(s ...string) (string, error), args ...string) error {
	tick := time.NewTicker(time.Millisecond * 1500)
	defer tick.Stop()

	var stat string
	var err error

	for _ = range tick.C {
		if len(args) > 0 {
			stat, err = getStatus(args[0])
		} else {
			stat, err = getStatus()
		}

		if err != nil {
			return err
		}
		switch stat {
		case
			"FAILED",
			"CREATE_COMPLETE",
			"DELETE_COMPLETE",
			"UPDATE_ROLLBACK_FAILED",
			"ROLLBACK_FAILED",
			"DELETE_FAILED",
			"CREATE_FAILED",
			"ROLLBACK_COMPLETE",
			"UPDATE_ROLLBACK_COMPLETE":
			return nil
		default:
			continue
		}
	}

	return nil
}
