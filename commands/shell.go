package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/daidokoro/qaz/utils"

	yaml "gopkg.in/yaml.v2"

	stks "github.com/daidokoro/qaz/stacks"

	"github.com/daidokoro/ishell"
	"github.com/spf13/cobra"
)

// define shell commands

var (
	shell = ishell.New()

	// define shell cmd
	shellCmd = &cobra.Command{
		Use:     "shell",
		Short:   "Qaz interactive shell - loads the specified config into an interactive shell",
		PreRun:  initialise,
		Example: "qaz shell -c config.yml",
		Run: func(cmd *cobra.Command, args []string) {
			// read config
			err := Configure(run.cfgSource, run.cfgRaw)
			utils.HandleError(err)

			// init shell
			initShell(config.Project, shell)

			// run shell
			shell.Run()
		},
	}
)

func initShell(p string, s *ishell.Shell) {
	var wg sync.WaitGroup
	// display welcome info.
	s.Println(fmt.Sprintf(
		"\n%s Shell Mode\n--\nTry \"help\" for a list of commands\n",
		log.ColorString("Qaz", "magenta"),
	))

	// arrary of commands
	shCommands := []*ishell.Cmd{
		// status command
		&ishell.Cmd{
			Name: "status",
			Help: "Prints status of deployed/un-deployed stacks",
			Func: func(c *ishell.Context) {
				var w sync.WaitGroup
				stacks.Range(func(k string, s *stks.Stack) bool {
					w.Add(1)
					go func() {
						defer w.Done()
						if err := s.Status(); err != nil {
							log.Error("failed to fetch status for [%s]: %v", s.Stackname, err)
						}
					}()
					return true
				})

				w.Wait()
				return
			},
		},

		// ls command
		&ishell.Cmd{
			Name: "ls",
			Help: "list all stacks defined in project config",
			Func: func(c *ishell.Context) {
				stacks.Range(func(k string, s *stks.Stack) bool {
					fmt.Println(k)
					return true
				})
			},
		},

		// outputs command
		&ishell.Cmd{
			Name:     "outputs",
			Help:     "Prints stack outputs",
			LongHelp: "outputs [stack]",
			Func: func(c *ishell.Context) {
				if len(c.Args) < 1 {
					log.Warn("please specify stack(s) to check")
					return
				}

				for _, s := range c.Args {
					// check if stack exists
					if _, ok := stacks.Get(s); !ok {
						log.Error("%s: does not exist in config", s)
						return
					}

					wg.Add(1)
					go func(s string) {
						defer wg.Done()
						if err := stacks.MustGet(s).Outputs(); err != nil {
							log.Error(err.Error())
							return
						}

						for _, i := range stacks.MustGet(s).Output.Stacks {
							m, err := json.MarshalIndent(i.Outputs, "", "  ")
							if err != nil {
								log.Error(err.Error())
							}

							reg, err := regexp.Compile(OutputRegex)
							utils.HandleError(err)

							resp := reg.ReplaceAllStringFunc(string(m), func(s string) string {
								return log.ColorString(s, "cyan")
							})

							fmt.Println(resp)
						}

						return
					}(s)
				}
				wg.Wait()
			},
		},

		// values command
		&ishell.Cmd{
			Name:     "values",
			Help:     "print stack values from config in YAML format",
			LongHelp: "values [stack]",
			Func: func(c *ishell.Context) {

				if len(c.Args) < 1 {
					log.Warn("please specify stack name...")
					return
				}

				// set stack value based on argument
				s := c.Args[0]

				if _, ok := stacks.Get(s); !ok {
					log.Error("stack [%s] not found in config", s)
					return
				}

				values := stacks.MustGet(s).TemplateValues[s].(map[string]interface{})

				log.Debug("converting stack outputs to JSON from: %s", values)
				output, err := yaml.Marshal(values)
				if err != nil {
					log.Error(err.Error())
					return
				}

				reg, err := regexp.Compile(".+?:(\n| )")
				if err != nil {
					log.Error(err.Error())
					return
				}

				resp := reg.ReplaceAllStringFunc(string(output), func(s string) string {
					return log.ColorString(s, "cyan")
				})

				fmt.Printf("\n%s\n", resp)
			},
		},

		// deploy command
		&ishell.Cmd{
			Name: "deploy",
			Help: "Deploys stack(s) to AWS",
			Func: func(c *ishell.Context) {
				// stack list
				var stklist []string
				stacks.Range(func(k string, _ *stks.Stack) bool {
					stklist = append(stklist, k)
					return true
				})

				// create checklist
				choices := c.Checklist(
					stklist,
					fmt.Sprintf("select stacks to %s:", log.ColorString("Deploy", "cyan")),
					nil,
				)

				// define actioned stacks
				for _, i := range choices {
					if i < 0 {
						fmt.Printf("--\nPress %s to return\n--\n", log.ColorString("ENTER", "green"))
						return
					}

					stacks.MustGet(stklist[i]).Actioned = true
				}

				// run actioned stacks
				stacks.Range(func(k string, s *stks.Stack) bool {
					if !s.Actioned {
						return true
					}

					if err := s.GenTimeParser(); err != nil {
						log.Error(err.Error())
						return false
					}
					return true
				})

				// Deploy Stacks
				stks.DeployHandler(&stacks)
				fmt.Printf("--\nPress %s to return\n--\n", log.ColorString("ENTER", "green"))
				return
			},
		},

		// terminate command
		&ishell.Cmd{
			Name: "terminate",
			Help: "Terminate stacks",
			Func: func(c *ishell.Context) {
				// stack list
				var stklist []string
				stacks.Range(func(k string, _ *stks.Stack) bool {
					stklist = append(stklist, k)
					return true
				})

				// create checklist
				choices := c.Checklist(
					stklist,
					fmt.Sprintf("select stacks to %s:", log.ColorString("Terminate", "red")),
					nil,
				)

				// define run.stacks
				for _, i := range choices {
					if i < 0 {
						fmt.Printf("--\nPress %s to return\n--\n", log.ColorString("ENTER", "green"))
						return
					}
					stacks.MustGet(stklist[i]).Actioned = true
				}

				// Terminate Stacks
				stks.TerminateHandler(&stacks)
				fmt.Printf("--\nPress %s to return\n--\n", log.ColorString("ENTER", "green"))
				return

			},
		},

		// generate command
		&ishell.Cmd{
			Name:     "generate",
			Help:     "generates template from configuration values",
			LongHelp: "generate [stack]",
			Func: func(c *ishell.Context) {
				var s string

				if len(c.Args) > 0 {
					s = c.Args[0]
				}

				// check if stack exists in config
				if _, ok := stacks.Get(s); !ok {
					log.Error("stack [%s] not found in config", s)
					return
				}

				if stacks.MustGet(s).Source == "" {
					log.Error("source not found in config file...")
					return
				}

				name := fmt.Sprintf("%s-%s", project, s)
				log.Debug("generating a template for [%s]", name)

				if err := stacks.MustGet(s).GenTimeParser(); err != nil {
					log.Error(err.Error())
					return
				}

				reg, err := regexp.Compile(OutputRegex)
				utils.HandleError(err)

				resp := reg.ReplaceAllStringFunc(string(stacks.MustGet(s).Template), func(s string) string {
					return log.ColorString(s, "cyan")
				})

				fmt.Println(resp)
			},
		},

		// check command
		&ishell.Cmd{
			Name:     "check",
			Help:     "validates cloudformation templates",
			LongHelp: "check [stack]",
			Func: func(c *ishell.Context) {
				var s string

				if len(c.Args) > 0 {
					s = c.Args[0]
				}

				// check if stack exists in config
				if _, ok := stacks.Get(s); !ok {
					log.Error("stack [%s] not found in config", s)
					return
				}

				if stacks.MustGet(s).Source == "" {
					log.Error("source not found in config file...")
					return
				}

				name := fmt.Sprintf("%s-%s", config.Project, s)
				log.Debug("validating template for %s", name)

				if err := stacks.MustGet(s).GenTimeParser(); err != nil {
					log.Error(err.Error())
				}

				if err := stacks.MustGet(s).Check(); err != nil {
					log.Error(err.Error())
				}
			},
		},

		// update command
		&ishell.Cmd{
			Name:     "update",
			Help:     "updates a given stack via change-set",
			LongHelp: "update [stack]",
			Func: func(c *ishell.Context) {
				var s string

				if len(c.Args) < 1 {
					log.Warn("please specify stack name...")
					return
				}

				// define stack name
				s = c.Args[0]

				// check if stack exists in config
				if _, ok := stacks.Get(s); !ok {
					log.Error("stack [%s] not found in config", s)
					return
				}

				if stacks.MustGet(s).Source == "" {
					log.Error("source not found in config file...")
					return
				}

				// random chcange-set name
				run.changeName = fmt.Sprintf(
					"%s-change-%s",
					stacks.MustGet(s).Stackname,
					strconv.Itoa((rand.Int())),
				)

				if err := stacks.MustGet(s).GenTimeParser(); err != nil {
					log.Error(err.Error())
					return
				}

				if err := stacks.MustGet(s).Change("create", run.changeName); err != nil {
					log.Error(err.Error())
					return
				}

				// descrupt change-set
				if err := stacks.MustGet(s).Change("desc", run.changeName); err != nil {
					log.Error(err.Error())
					return
				}

				for {
					c.Print(fmt.Sprintf(
						"--\n%s [%s]: ",
						log.ColorString("The above will be updated, do you want to proceed?", "red"),
						log.ColorString("Y/N", "cyan"),
					))

					resp := c.ReadLine()
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
			},
		},

		// set-policy command
		&ishell.Cmd{
			Name:     "set-policy",
			Help:     "set stack policies based on configured value",
			LongHelp: "set-policy [stack]",
			Func: func(c *ishell.Context) {
				var wg sync.WaitGroup
				// stack list
				stklist := make([]string, stacks.Count())
				stacks.Range(func(k string, _ *stks.Stack) bool {
					stklist = append(stklist, k)
					return true
				})

				// create checklist
				choices := c.Checklist(
					stklist,
					fmt.Sprintf("select stacks to %s:", log.ColorString("set-policy", "yellow")),
					nil,
				)

				// define run.stacks
				for _, i := range choices {
					if i < 0 {
						fmt.Printf("--\nPress %s to return\n--\n", log.ColorString("ENTER", "green"))
						return
					}
					stacks.MustGet(stklist[i]).Actioned = true
				}

				stacks.Range(func(k string, s *stks.Stack) bool {
					if !s.Actioned {
						return true
					}

					wg.Add(1)
					go func() {
						defer wg.Done()
						if err := s.StackPolicy(); err != nil {
							log.Error("%v", err)
							return
						}
						return
					}()
					return true
				})

				wg.Wait()

			},
		},

		// reload command
		&ishell.Cmd{
			Name:     "reload",
			Help:     "reload Qaz configuration source into shell environment",
			LongHelp: "reload",
			Func: func(c *ishell.Context) {
				// off load stacks by redeclaring stack map
				stacks = stks.Map{}

				// re-read config
				err := Configure(run.cfgSource, run.cfgRaw)
				utils.HandleError(err)
				log.Info("config reloaded: [%s]", run.cfgSource)
			},
		},

		// lint command
		&ishell.Cmd{
			Name:     "lint",
			Help:     "lints template using cfn-lint",
			LongHelp: "lint [stack]",
			Func: func(c *ishell.Context) {
				var s string

				if len(c.Args) > 0 {
					s = c.Args[0]
				}

				// check if stack exists in config
				if _, ok := stacks.Get(s); !ok {
					log.Error("stack [%s] not found in config", s)
					return
				}

				if stacks.MustGet(s).Source == "" {
					log.Error("source not found in config file...")
					return
				}

				name := fmt.Sprintf("%s-%s", project, s)
				log.Debug("generating a template for [%s]", name)

				if err := stacks.MustGet(s).GenTimeParser(); err != nil {
					log.Error(err.Error())
					return
				}

				// write template to temporary file
				content := []byte(stacks.MustGet(s).Template)
				filename := fmt.Sprintf(".%s.qaz", s)
				writeErr := ioutil.WriteFile(filename, content, 0644)
			  if writeErr != nil {
					log.Error(writeErr.Error())
					return
				}

				// run cfn-lint against temporary file
				_, lookErr := exec.LookPath("cfn-lint")
				if lookErr != nil {
					log.Error("cfn-lint executable not found! Please consider https://pypi.org/project/cfn-lint/ for help.")
					return
				}
				execCmd := exec.Command("cfn-lint", filename)
				execCmd.Env = append(os.Environ())
				execCmd.Stdout = os.Stdout
				execCmd.Stderr = os.Stderr
				execErr := execCmd.Run()
				if execErr != nil {
					log.Error(execErr.Error())
					return
				}

			},
		},

	}

	// set prompt
	s.SetPrompt(fmt.Sprintf(
		"%s %s:(%s) %s ",
		log.ColorString("@", "yellow"),
		log.ColorString("qaz", "cyan"),
		log.ColorString(p, "magenta"),
		log.ColorString("✗", "green"),
	))

	// add commands
	for _, c := range shCommands {
		s.AddCmd(c)
	}
}
