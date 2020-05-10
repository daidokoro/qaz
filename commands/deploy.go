package commands

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/daidokoro/qaz/repo"
	"github.com/daidokoro/qaz/utils"

	stks "github.com/daidokoro/qaz/stacks"

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

			// Add run.stacks based on [templates] Flags
			for _, src := range run.tplSources {
				s, source, err := utils.GetSource(src)
				utils.HandleError(err)
				if _, ok := stacks.Get(s); !ok {
					utils.HandleError(fmt.Errorf("stacks [%s] not found in config", s))
				}
				stacks.MustGet(s).Source = source
				stacks.MustGet(s).Actioned = true
			}

			// Add all stacks with defined sources if actioned
			if run.all {
				stacks.Range(func(_ string, s *stks.Stack) bool {
					s.Actioned = true
					return true
				})
			}

			// Add run.stacks based on Args
			if len(args) > 0 && !run.all {
				for _, s := range args {
					if _, ok := stacks.Get(s); !ok {
						utils.HandleError(fmt.Errorf("stacks [%s] not found in config", s))
					}
					stacks.MustGet(s).Actioned = true
				}
			}

			// run gentimeParser
			stacks.Range(func(_ string, s *stks.Stack) bool {
				if !s.Actioned {
					return true
				}
				if err := s.GenTimeParser(); err != nil {
					utils.HandleError(err)
				}
				return true
			})

			// Deploy Stacks
			stks.DeployHandler(&stacks)

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

			repo, err := repo.NewRepo(args[0], run.gituser, run.gitrsa)
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

			//create set actioned stacks
			stacks.Range(func(_ string, s *stks.Stack) bool {
				s.Actioned = true
				utils.HandleError(s.GenTimeParser())
				return true
			})

			// Deploy Stacks
			stks.DeployHandler(&stacks)

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

			switch {

			case run.tplSource != "":
				s, source, err = utils.GetSource(run.tplSource)
				utils.HandleError(err)

			case len(args) > 0:
				s = args[0]
				if _, ok := stacks.Get(s); !ok {
					utils.HandleError(fmt.Errorf("stacks [%s] not found in config", s))
				}
			}

			// check stack exists
			if _, ok := stacks.Get(s); !ok {
				utils.HandleError(fmt.Errorf("stacks [%s] not found in config", s))
			}

			if source != "" {
				stacks.MustGet(s).Source = source
			}

			utils.HandleError(stacks.MustGet(s).GenTimeParser())

			if run.interactive {
				// random change-set name
				run.changeName = fmt.Sprintf(
					"%s-change-%s",
					stacks.MustGet(s).Stackname,
					strconv.Itoa((rand.Int())),
				)

				if err := stacks.MustGet(s).Change("create", run.changeName); err != nil {
					log.Error(err.Error())
					return
				}

				// describe change-set
				if err := stacks.MustGet(s).Change("desc", run.changeName); err != nil {
					log.Error(err.Error())
					return
				}

				for {
					fmt.Println(fmt.Sprintf(
						"--\n%s [%s]: ",
						log.ColorString("The above will be updated, do you want to proceed?", "red"),
						log.ColorString("Y/N", "cyan"),
					))

					scanner := bufio.NewScanner(os.Stdin)
					scanner.Scan()
					resp := scanner.Text()
					switch strings.ToLower(resp) {
					case "y":
						if err := stacks.MustGet(s).Change("execute", run.changeName); err != nil {
							log.Error(err.Error())
							return
						}
						log.Info("update completed successfully...")
						return
					case "n":
						if err := stacks.MustGet(s).Change("rm", run.changeName); err != nil {
							log.Error(err.Error())
							return
						}
						return
					default:
						log.Warn(`invalid response, please type "Y" or "N"`)
						continue
					}
				}

			} else {
				// non-interactive mode
				utils.HandleError(stacks.MustGet(s).Update())
			}

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

			err := Configure(run.cfgSource, "")
			utils.HandleError(err)

			// select actioned stacks
			for _, s := range args {
				if _, ok := stacks.Get(s); !ok {
					utils.HandleError(fmt.Errorf("stacks [%s] not found in config", s))
				}
				stacks.MustGet(s).Actioned = true
			}

			// action stacks if all
			if run.all {
				stacks.Range(func(_ string, s *stks.Stack) bool {
					s.Actioned = true
					return true
				})
			}

			// Terminate Stacks
			stks.TerminateHandler(&stacks)
		},
	}
)
