package stacks

import "fmt"

// support for SAM - Serverless Arch Model Cloudformation templates

// DeployServerless deploys SAMs Cloudformation templates
func (s *Stack) DeployServerless() error {
	changename := fmt.Sprintf("%s-change-set", s.Stackname)

	Log.Debug("deploying serverless template via change-set")
	if err := s.Change("transform", changename); err != nil {
		return err
	}

	if err := s.Change("execute", changename); err != nil {
		return err
	}

	return nil
}
