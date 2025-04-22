go
package cmdSessions

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/daidokoro/qaz/log"
	"github.com/spf13/cobra"
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

var (
	run = struct {
		cfgSource   string
		tplSource   string
		profile     string
		region      string
		tplSources  []string
		stacks      map[string]string
		all         bool
		version     bool
		request     string
		debug       bool
		funcEvent   string
		lambdAsync  bool
		changeName  string
		stackName   string
		rollback    bool
		colors      bool
		cfgRaw      string
		gituser     string
		gitpass     string
		gitrsa      string
		protectOff  bool
		interactive bool
	}{}
)

// NewCmdSessions - command to interact with AWS Sessions
func NewCmdSessions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "Manage AWS sessions",
		Long:  `Manage AWS sessions`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("sessions called")
		},
	}
	return cmd
}


