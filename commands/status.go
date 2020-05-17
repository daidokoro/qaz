package commands

import (
	"fmt"
	"strings"
	"sync"

	"github.com/daidokoro/qaz/log"
	"github.com/daidokoro/qaz/repo"
	"github.com/daidokoro/qaz/utils"

	"github.com/daidokoro/qaz/stacks"

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
			stks, err := Configure(run.cfgSource, run.cfgRaw)
			utils.HandleError(err)

			stks.Range(func(_ string, s *stacks.Stack) bool {
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

	// git-status command
	gitStatusCmd = &cobra.Command{
		Use:     "git-status [git-repo]",
		Short:   "Check status of deployment via files stored in Git repository",
		Example: "qaz git-status https://github.com/cfn-deployable/simplevpc --user me",
		PreRun:  initialise,
		Run: func(cmd *cobra.Command, args []string) {

			// check args
			if len(args) < 1 {
				fmt.Println("Please specify git repo...")
				return
			}

			var wg sync.WaitGroup

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

			stks.Range(func(_ string, s *stacks.Stack) bool {
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

			stks, err := Configure(run.cfgSource, "")
			utils.HandleError(err)

			switch {
			case run.tplSource != "":
				s, source, err = utils.GetSource(run.tplSource)
				utils.HandleError(err)
			case len(args) > 0:
				s = args[0]
			}

			// check if stack exists in config
			if _, ok := stks.Get(s); !ok {
				utils.HandleError(fmt.Errorf("stack [%s] not found in config", s))
			}

			if source != "" {
				stks.MustGet(s).Source = source
			}

			name := fmt.Sprintf("%s-%s", config.Project, s)
			log.Info("validating template: %s", name)

			utils.HandleError(stks.MustGet(s).GenTimeParser())
			utils.HandleError(stks.MustGet(s).Check())
		},
	}

	// protect command
	protectCmd = &cobra.Command{
		Use:    "protect",
		Short:  "Enables stack termination protection",
		PreRun: initialise,
		Example: strings.Join([]string{
			"qaz protect --all",
			"qaz protect --all --off",
			"qaz protect stack-name --off",
		}, "\n"),
		Run: func(cmd *cobra.Command, args []string) {

			req := "enabled"
			if run.protectOff {
				req = "disabled"
			}

			if len(args) < 1 && !run.all {
				log.Warn("No stack specified for termination")
				return
			}

			stks, err := Configure(run.cfgSource, "")
			utils.HandleError(err)

			var w sync.WaitGroup

			// if run all
			if run.all {
				stks.Range(func(_ string, s *stacks.Stack) bool {
					w.Add(1)
					go func() {
						defer w.Done()
						if err := s.Protect(&run.protectOff); err != nil {
							log.Error("error enable termination-protection on: [%s] - %v", s.Name, err)
							return
						}
						log.Info("termination protection %s: [%s]", req, s.Name)
					}()
					return true
				})
				w.Wait()
				return
			}

			// if individual
			for _, s := range args {
				if _, ok := stks.Get(s); !ok {
					utils.HandleError(fmt.Errorf("stacks [%s] not found in config", s))
				}
				w.Add(1)
				go func(s string) {
					defer w.Done()
					if err := stks.MustGet(s).Protect(&run.protectOff); err != nil {
						log.Error("error enable termination-protection on: [%s] - %v", s, err)
						return
					}
					log.Info("termination protection %s: [%s]", req, s)
				}(s)
			}
			w.Wait()

		},
	}
)
