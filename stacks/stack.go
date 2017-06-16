package stacks

import (
	"fmt"
	"strings"
	"sync"

	"github.com/daidokoro/qaz/logger"
	"github.com/daidokoro/qaz/repo"

	"text/template"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

var (
	// Log defines Logger
	Log *logger.Logger

	// define waitGroup
	wg sync.WaitGroup

	// Git repo for stack deployment
	Git *repo.Repo
)

// Stack - holds all meaningful information about a particular stack.
type Stack struct {
	Project        *string
	Name           string
	Stackname      string
	Template       string
	DependsOn      []string
	Dependents     []interface{}
	Stackoutputs   *cloudformation.DescribeStacksOutput
	Parameters     []*cloudformation.Parameter
	Output         *cloudformation.DescribeStacksOutput
	Policy         string
	Tags           []*cloudformation.Tag
	Session        *session.Session
	Profile        string
	Source         string
	Bucket         string
	Role           string
	Rollback       bool
	GenTimeFunc    *template.FuncMap
	DeployTimeFunc *template.FuncMap
	DeployDelims   *string
	GenDelims      *string
	TemplateValues map[string]interface{}
	Debug          bool
	Timeout        int64
}

// SetStackName - sets the stackname with struct
func (s *Stack) SetStackName() {
	s.Stackname = fmt.Sprintf("%s-%s", *s.Project, s.Name)
}

// creds - Returns credentials if role set
func (s *Stack) creds() *credentials.Credentials {
	var creds *credentials.Credentials
	if s.Role == "" {
		return creds
	}
	return stscreds.NewCredentials(s.Session, s.Role)
}

// delims - returns delimiters for parsing templates
func (s *Stack) delims(lvl string) (string, string) {
	if lvl == "deploy" {
		if *s.DeployDelims != "" {
			delims := strings.Split(*s.DeployDelims, ":")
			return delims[0], delims[1]
		}

		// default
		return "<<", ">>"
	}

	if *s.GenDelims != "" {
		delims := strings.Split(*s.GenDelims, ":")
		return delims[0], delims[1]
	}

	// default
	return "{{", "}}"
}
