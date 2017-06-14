package commands

import (
	"fmt"
	"qaz/repo"
	"qaz/utils"
	"strings"

	stks "qaz/stacks"

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

			err := Configure(run.cfgSource, run.cfgRaw)
			utils.HandleError(err)

			run.stacks = make(map[string]string)

			// Add run.stacks based on templates Flags
			for _, src := range run.tplSources {
				s, source, err := utils.GetSource(src)
				utils.HandleError(err)
				run.stacks[s] = source
			}

			// Add all stacks with defined sources if all
			if run.all {
				for s, v := range stacks {
					// so flag values aren't overwritten
					if _, ok := run.stacks[s]; !ok {
						run.stacks[s] = v.Source
					}
				}
			}

			// Add run.stacks based on Args
			if len(args) > 0 && !run.all {
				for _, stk := range args {
					if _, ok := stacks[stk]; !ok {
						utils.HandleError(fmt.Errorf("Stack [%s] not found in conig", stk))
					}
					run.stacks[stk] = stacks[stk].Source
				}
			}

			for s, src := range run.stacks {
				if stacks[s].Source == "" {
					stacks[s].Source = src
				}

				err := stacks[s].GenTimeParser()
				utils.HandleError(err)

				// Handle missing stacks
				if stacks[s] == nil {
					utils.HandleError(fmt.Errorf("Missing Stack in %s: [%s]", run.cfgSource, s))
				}
			}

			// Deploy Stacks
			stks.DeployHandler(run.stacks, stacks)

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

			repo, err := repo.NewRepo(args[0])
			utils.HandleError(err)

			// Passing repo to the global var
			gitrepo = *repo

			// add repo
			stks.Git = &gitrepo

			if out, ok := repo.Files[run.cfgSource]; ok {
				repo.Config = out
			}

			log.Debug("Repo Files:")
			for k := range repo.Files {
				log.Debug(k)
			}

			err = Configure(run.cfgSource, repo.Config)
			utils.HandleError(err)

			//create run stacks
			run.stacks = make(map[string]string)

			for s, v := range stacks {
				// populate run stacks
				run.stacks[s] = v.Source
				err := stacks[s].GenTimeParser()
				utils.HandleError(err)
			}

			// Deploy Stacks
			stks.DeployHandler(run.stacks, stacks)

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

			err := Configure(run.cfgSource, run.cfgRaw)
			if err != nil {
				utils.HandleError(err)
				return
			}

			if run.tplSource != "" {
				s, source, err = utils.GetSource(run.tplSource)
				if err != nil {
					utils.HandleError(err)
					return
				}
			}

			if len(args) > 0 {
				s = args[0]
			}

			// check if stack exists in config
			if _, ok := stacks[s]; !ok {
				utils.HandleError(fmt.Errorf("Stack [%s] not found in config", s))
				return
			}

			if stacks[s].Source == "" {
				stacks[s].Source = source
			}

			err = stacks[s].GenTimeParser()
			if err != nil {
				utils.HandleError(err)
				return
			}

			// Handle missing stacks
			if stacks[s] == nil {
				utils.HandleError(fmt.Errorf("Missing Stack in %s: [%s]", run.cfgSource, s))
				return
			}

			if err := stacks[s].Update(); err != nil {
				utils.HandleError(err)
				return
			}

		},
	}

	// terminate command
	terminateCmd = &cobra.Command{
		Use:    "terminate [stacks]",
		Short:  "Terminates stacks",
		PreRun: initialise,
		Run: func(cmd *cobra.Command, args []string) {

			if !run.all {
				run.stacks = make(map[string]string)
				for _, stk := range args {
					run.stacks[stk] = ""
				}

				if len(run.stacks) == 0 {
					log.Warn("No stack specified for termination")
					return
				}
			}

			err := Configure(run.cfgSource, "")
			if err != nil {
				utils.HandleError(err)
				return
			}

			// Terminate Stacks
			stks.TerminateHandler(run.stacks, stacks)
		},
	}
)
