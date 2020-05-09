package commands

import (
	"fmt"
	"strings"

	"github.com/daidokoro/qaz/log"
	"github.com/daidokoro/qaz/repo"
	"github.com/daidokoro/qaz/stacks"
	"github.com/daidokoro/qaz/utils"

	"github.com/spf13/cobra"
)

// stack management commands, ie. deploy, terminate, update

var (
	// deploy command
	deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploys stack(s) to AWS",
		Example: strings.Join([]string{
			"qaz deploy stack -c path/to/config",
			"qaz deploy -c path/to/config -t stack::s3://bucket/key",
			"qaz deploy -c path/to/config -t stack::path/to/template",
			"qaz deploy -c path/to/config -t stack::http://someurl",
			"qaz deploy -c path/to/config -t stack::lambda:{some:json}@lambda_function",
		}, "\n"),
		PreRun: initialise,
		Run: func(cmd *cobra.Command, args []string) {

			stks, err := Configure(run.cfgSource, run.cfgRaw)
			utils.HandleError(err)

			run.stacks = make(map[string]string)

			// Add run.stacks based on [templates] Flags
			for _, src := range run.tplSources {
				s, source, err := utils.GetSource(src)
				utils.HandleError(err)
				if _, ok := stks.Get(s); !ok {
					utils.HandleError(fmt.Errorf("stacks [%s] not found in config", s))
				}
				stks.MustGet(s).Source = source
				stks.MustGet(s).Actioned = true
			}

			// Add all stacks with defined sources if actioned
			if run.all {
				stks.Range(func(_ string, s *stacks.Stack) bool {
					s.Actioned = true
					return true
				})
			}

			// Add run.stacks based on Args
			if len(args) > 0 && !run.all {
				for _, s := range args {
					if _, ok := stks.Get(s); !ok {
						utils.HandleError(fmt.Errorf("stacks [%s] not found in config", s))
					}
					stks.MustGet(s).Actioned = true
				}
			}

			// run gentimeParser
			stks.Range(func(_ string, s *stacks.Stack) bool {
				if !s.Actioned {
					return true
				}
				if err := s.GenTimeParser(); err != nil {
					utils.HandleError(err)
				}
				return true
			})

			// Deploy Stacks
			stacks.DeployHandler(&stks)

		},
	}

	// git-deploy command
	gitDeployCmd = &cobra.Command{
		Use:     "git-deploy [git-repo]",
		Short:   "Deploy project from Git repository",
		Example: "qaz git-deploy https://github.com/cfn-deployable/simplevpc --user me",
		PreRun:  initialise,
		Run: func(cmd *cobra.Command, args []string) {

			// check args
			if len(args) < 1 {
				fmt.Println("Please specify git repo...")
				return
			}

			repo, err := repo.New(args[0], run.gituser, run.gitrsa)
			utils.HandleError(err)

			// Passing repo to the global var
			gitrepo = *repo

			// add repo
			stacks.Git(&gitrepo)

			if out, ok := repo.Files[run.cfgSource]; ok {
				repo.Config = out
			}

			log.Debug("Repo Files:")
			for k := range repo.Files {
				log.Debug(k)
			}

			stks, err := Configure(run.cfgSource, repo.Config)
			utils.HandleError(err)

			//create set actioned stacks
			stks.Range(func(_ string, s *stacks.Stack) bool {
				s.Actioned = true
				utils.HandleError(s.GenTimeParser())
				return true
			})

			// Deploy Stacks
			stacks.DeployHandler(&stks)
		},
	}

	// update command
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Updates a given stack",
		Example: strings.Join([]string{
			"qaz update -c path/to/config -t stack::path/to/template",
			"qaz update -c path/to/config -t stack::s3://bucket/key",
			"qaz update -c path/to/config -t stack::http://someurl",
			"qaz deploy -c path/to/config -t stack::lambda:{some:json}@lambda_function",
		}, "\n"),
		PreRun: initialise,
		Run: func(cmd *cobra.Command, args []string) {

			var s string
			var source string

			stks, err := Configure(run.cfgSource, run.cfgRaw)
			if err != nil {
				utils.HandleError(err)
				return
			}

			switch {

			case run.tplSource != "":
				s, source, err = utils.GetSource(run.tplSource)
				utils.HandleError(err)

			case len(args) > 0:
				s = args[0]
				if _, ok := stks.Get(s); !ok {
					utils.HandleError(fmt.Errorf("stacks [%s] not found in config", s))
				}
			}

			// check stack exists
			if _, ok := stks.Get(s); !ok {
				utils.HandleError(fmt.Errorf("stacks [%s] not found in config", s))
			}

			if source != "" {
				stks.MustGet(s).Source = source
			}

			utils.HandleError(stks.MustGet(s).GenTimeParser())
			utils.HandleError(stks.MustGet(s).Update())
		},
	}

	// terminate command
	terminateCmd = &cobra.Command{
		Use:    "terminate [stacks]",
		Short:  "Terminates stacks",
		PreRun: initialise,
		Run: func(cmd *cobra.Command, args []string) {

			if len(args) < 1 && !run.all {
				log.Warn("No stack specified for termination")
				return
			}

			stks, err := Configure(run.cfgSource, "")
			utils.HandleError(err)

			// select actioned stacks
			for _, s := range args {
				if _, ok := stks.Get(s); !ok {
					utils.HandleError(fmt.Errorf("stacks [%s] not found in config", s))
				}
				stks.MustGet(s).Actioned = true
			}

			// action stacks if all
			if run.all {
				stks.Range(func(_ string, s *stacks.Stack) bool {
					s.Actioned = true
					return true
				})
			}

			// Terminate Stacks
			stacks.TerminateHandler(&stks)
		},
	}
)
