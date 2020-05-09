// Package testing contains all unit test for the qaz code base
package testing

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/daidokoro/qaz/log"
)

var (
	// aws session
	sess = session.Must(session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1"),
	}))
)

// TestMain - test initialisation
func TestMain(m *testing.M) {
	// configure logging
	log.SetDefault(log.NewDefaultLogger(true, true))
	ecode := m.Run()
	os.Exit(ecode)
}
