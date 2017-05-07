package commands

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// SessionManager - handles AWS Sessions
type sessionManager struct {
	region   string
	sessions map[string]*session.Session
}

// GetSess - Returns aws session based on given profile
func (s *sessionManager) GetSess(p string) (*session.Session, error) {

	var sess *session.Session

	// Set P to default or command input if stack input is empty
	if p == "" {
		p = job.profile
	}

	if v, ok := s.sessions[p]; ok {
		Log(fmt.Sprintf("Session Detected: [%s]", p), level.debug)
		return v, nil
	}

	options := session.Options{
		Profile:           p,
		SharedConfigState: session.SharedConfigEnable,
	}

	if s.region != "" {
		options.Config = aws.Config{Region: &s.region}
	}

	Log(fmt.Sprintf("Creating AWS Session with options: Regioin: %s, Profile: %s ", region, job.profile), level.debug)
	sess, err := session.NewSessionWithOptions(options)
	if err != nil {
		return sess, err
	}

	s.sessions[p] = sess
	return sess, nil
}

var manager = sessionManager{
	sessions: make(map[string]*session.Session),
}
