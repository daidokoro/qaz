package commands

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/daidokoro/qaz/log"
)

// GetSession - Returns aws session based on default run.profile and run.Region
// options can be overwritten by passing a function: func(*session.Options)
func GetSession(options ...func(*session.Options)) (*session.Session, error) {

	var sess *session.Session

	// set default config values
	opts := &session.Options{
		Profile:                 run.profile,
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
	}

	if run.region != "" {
		opts.Config = aws.Config{Region: &run.region}
	}

	// apply external Options
	for _, f := range options {
		f(opts)
	}

	log.Debug("Creating AWS Session with options: %s", opts)
	sess, err := session.NewSessionWithOptions(*opts)
	if err != nil {
		return sess, err
	}

	return sess, nil
}
