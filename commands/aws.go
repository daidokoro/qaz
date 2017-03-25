package commands

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// Declaring single global session.
var conn *session.Session
var once sync.Once

func awsSession() (*session.Session, error) {
	var err error

	// Using sync.Once to ensure session is created only once.
	once.Do(func() {
		// set region
		var r string
		switch config.Region {
		case "":
			r = region
		default:
			r = config.Region
		}

		//define session options
		options := session.Options{
			Config:            aws.Config{Region: &r},
			Profile:           job.profile,
			SharedConfigState: session.SharedConfigEnable,
		}

		Log(fmt.Sprintf("Creating AWS Session with options: Regioin: %s, Profile: %s ", region, job.profile), level.debug)
		conn, err = session.NewSessionWithOptions(options)
	})

	if err != nil {
		return conn, err
	}

	return conn, nil
}
