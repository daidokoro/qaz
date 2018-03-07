package commands

import (
	"fmt"
	"strings"
	"sync"

	"github.com/daidokoro/qaz/utils"

	stks "github.com/daidokoro/qaz/stacks"

	"github.com/spf13/cobra"
)

// status and validation based commands

var (
	// status command
	statusCmd = &cobra.Command{
		Use:    "status",
		Short:  "Prints status of deployed/un-deployed stacks",
		PreRun: initialise,
		Run: func(cmd *cobra.Command, args []string) {
			var wg sync.WaitGroup
			err := Configure(run.cfgSource, run.cfgRaw)
			utils.HandleError(err)

			stacks.Range(func(_ string, s *stks.Stack) bool {
				wg.Add(1)
				go func() {
					if err := s.Status(); err != nil {
						log.Error("failed to fetch status for [%s]: %v", s.Stackname, err)
					}
					wg.Done()
				}()
				return true
			})

			wg.Wait()
		},
	}

	// validate/check command
	checkCmd = &cobra.Command{
		Use:   "check",
		Short: "Validates Cloudformation Templates",
		Example: strings.Join([]string{
			"qaz check -c path/to/config.yml -t path/to/template -c path/to/config",
			"qaz check -c path/to/config.yml -t stack::http://someurl",
			"qaz check -c path/to/config.yml -t stack::s3://bucket/key",
			"qaz deploy -c path/to/config.yml -t stack::lambda:{some:json}@lambda_function",
		}, "\n"),
		PreRun: initialise,
		Run: func(cmd *cobra.Command, args []string) {

			var s string
			var source string

			err := Configure(run.cfgSource, "")
			utils.HandleError(err)

			switch {
			case run.tplSource != "":
				s, source, err = utils.GetSource(run.tplSource)
				utils.HandleError(err)
			case len(args) > 0:
				s = args[0]
			}

			// check if stack exists in config
			if _, ok := stacks.Get(s); !ok {
				utils.HandleError(fmt.Errorf("stack [%s] not found in config", s))
			}

			if source != "" {
				stacks.MustGet(s).Source = source
			}

			name := fmt.Sprintf("%s-%s", config.Project, s)
			log.Info("validating template: %s", name)

			utils.HandleError(stacks.MustGet(s).GenTimeParser())
			utils.HandleError(stacks.MustGet(s).Check())
		},
	}
)
