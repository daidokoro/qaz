package stacks

import (
	"fmt"
	"strings"
	"sync"

	"github.com/daidokoro/qaz/repo"

	"text/template"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

var (
	// define waitGroup
	wg sync.WaitGroup

	// Git repo for stack deployment
	gitrepo *repo.Repo

	// OutputRegex for printing yaml/json output
	OutputRegex = `(?m)^[ -]*([^\r\n :]+?)\s*:`
)

// Stack - holds all meaningful information about a particular stack.
type Stack struct {
	// Project name for stack
	Project *string

	// local logical stack name
	Name string

	// AWS stack name on build
	Stackname string

	// Stack template for generating CF
	Template string

	// list of Dependencies
	DependsOn      []string
	Dependents     []interface{}
	Stackoutputs   *cloudformation.DescribeStacksOutput
	Parameters     []*cloudformation.Parameter
	Output         *cloudformation.DescribeStacksOutput
	Policy         string
	Tags           []*cloudformation.Tag
	Session        *session.Session
	Profile        string
	Region         string
	Source         string
	Bucket         string
	Role           string
	Rollback       bool
	GenTimeFunc    *template.FuncMap
	DeployTimeFunc *template.FuncMap
	DeployDelims   *string
	GenDelims      *string
	TemplateValues map[string]interface{}

	// Debug value
	// Debug   bool
	Timeout int64

	// Actioned in this context means the stack name
	// has been passed explicitly as an arguement and
	// should be processed
	Actioned bool

	// list of SNS notification ARNs
	NotificationARNs []string
}

// SetStackName - sets the.Stackname with struct
func (s *Stack) SetStackName() {
	if s.Stackname == "" {
		s.Stackname = fmt.Sprintf("%s-%s", *s.Project, s.Name)
	}
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

// Git - set git repo for stack deployments
func Git(r *repo.Repo) {
	gitrepo = r
}
