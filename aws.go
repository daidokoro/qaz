package main

import (
	"errors"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// Declaring single global session.
var sess *session.Session
var once sync.Once

func awsSession() (*session.Session, error) {
	var err error

	// Using sync.Once to ensure session is created only once.
	once.Do(func() {
		//define session options
		options := session.Options{
			Config:            aws.Config{Region: aws.String(region)},
			Profile:           job.profile,
			SharedConfigState: session.SharedConfigEnable,
		}

		Log(fmt.Sprintf("Creating AWS Session with options: Regioin: %s, Profile: %s ", region, job.profile), level.debug)
		sess, err = session.NewSessionWithOptions(options)
	})

	if err != nil {
		return &session.Session{}, errors.New(fmt.Sprintln("Failed establishing AWS session ", err))
	}

	return sess, nil
}
