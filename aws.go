// various small aws helper funcitons in this file
package main

import (
	"log"
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
		sess, err = session.NewSessionWithOptions(session.Options{
			Config:            aws.Config{Region: aws.String(region)},
			Profile:           job.profile,
			SharedConfigState: session.SharedConfigEnable,
		})
	})

	if err != nil {
		log.Fatal("Failed establishing AWS session ", err)
		return &session.Session{}, err
	}

	return sess, nil
}

// helper functions for finding ids of AWS resources
func subnetFinder(s string) (string, error) { return "", nil }
func sgFinder(s string) (string, error)     { return "", nil }
func amiFinder(s string) (string, error)    { return "", nil }
func vpcFinder(s string) (string, error)    { return "", nil }
