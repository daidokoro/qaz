package commands

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

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
				for _, v := range stacks {
					wg.Add(1)
					go func(s *stks.Stack) {
						if err := s.Status(); err != nil {
							log.Error("failed to fetch status for [%s]: %v", s.Stackname, err)
						}
						wg.Done()
					}(v)

				}
				wg.Wait()
			},
		},

		// ls command
		&ishell.Cmd{
			Name: "ls",
			Help: "list all stacks defined in project config",
			Func: func(c *ishell.Context) {
				for k := range stacks {
					fmt.Println(k)
				}
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
					if _, ok := stacks[s]; !ok {
						log.Error("%s: does not exist in config", s)
						return
					}

					wg.Add(1)
					go func(s string) {
						defer wg.Done()
						if err := stacks[s].Outputs(); err != nil {
							log.Error(err.Error())
							return
						}

						for _, i := range stacks[s].Output.Stacks {
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

				if _, ok := stacks[s]; !ok {
					log.Error("stack [%s] not found in config", s)
					return
				}

				values := stacks[s].TemplateValues[s].(map[string]interface{})

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
				run.stacks = make(map[string]string)
				// stack list
				stklist := make([]string, len(stacks))
				i := 0
				for k := range stacks {
					stklist[i] = k
					i++
				}

				// create checklist
				choices := c.Checklist(
					stklist,
					fmt.Sprintf("select stacks to %s:", log.ColorString("Deploy", "cyan")),
					nil,
				)

				// define run.stacks
				run.stacks = make(map[string]string)
				for _, i := range choices {
					if i < 0 {
						fmt.Printf("--\nPress %s to return\n--\n", log.ColorString("ENTER", "green"))
						return
					}
					run.stacks[stklist[i]] = ""
				}

				for s := range run.stacks {
					if err := stacks[s].GenTimeParser(); err != nil {
						log.Error(err.Error())
						return
					}
				}

				// Deploy Stacks
				stks.DeployHandler(run.stacks, stacks)
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
				stklist := make([]string, len(stacks))
				i := 0
				for k := range stacks {
					stklist[i] = k
					i++
				}

				// create checklist
				choices := c.Checklist(
					stklist,
					fmt.Sprintf("select stacks to %s:", log.ColorString("Terminate", "red")),
					nil,
				)

				// define run.stacks
				run.stacks = make(map[string]string)
				for _, i := range choices {
					if i < 0 {
						fmt.Printf("--\nPress %s to return\n--\n", log.ColorString("ENTER", "green"))
						return
					}
					run.stacks[stklist[i]] = ""
				}

				// Terminate Stacks
				stks.TerminateHandler(run.stacks, stacks)
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
				if _, ok := stacks[s]; !ok {
					log.Error("stack [%s] not found in config", s)
					return
				}

				if stacks[s].Source == "" {
					log.Error("source not found in config file...")
					return
				}

				name := fmt.Sprintf("%s-%s", project, s)
				log.Debug("generating a template for [%s]", name)

				if err := stacks[s].GenTimeParser(); err != nil {
					log.Error(err.Error())
					return
				}

				reg, err := regexp.Compile(OutputRegex)
				utils.HandleError(err)

				resp := reg.ReplaceAllStringFunc(string(stacks[s].Template), func(s string) string {
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
				if _, ok := stacks[s]; !ok {
					log.Error("stack [%s] not found in config", s)
					return
				}

				if stacks[s].Source == "" {
					log.Error("source not found in config file...")
					return
				}

				name := fmt.Sprintf("%s-%s", config.Project, s)
				log.Debug("validating template for %s", name)

				if err := stacks[s].GenTimeParser(); err != nil {
					log.Error(err.Error())
				}

				if err := stacks[s].Check(); err != nil {
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
				if _, ok := stacks[s]; !ok {
					log.Error("stack [%s] not found in config", s)
					return
				}

				if stacks[s].Source == "" {
					log.Error("source not found in config file...")
					return
				}

				// random chcange-set name
				run.changeName = fmt.Sprintf(
					"%s-change-%s",
					stacks[s].Stackname,
					strconv.Itoa((rand.Int())),
				)

				if err := stacks[s].GenTimeParser(); err != nil {
					log.Error(err.Error())
					return
				}

				if err := stacks[s].Change("create", run.changeName); err != nil {
					log.Error(err.Error())
					return
				}

				// descrupt change-set
				if err := stacks[s].Change("desc", run.changeName); err != nil {
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
						if err := stacks[s].Change("execute", run.changeName); err != nil {
							log.Error(err.Error())
							return
						}
						log.Info("update completed successfully...")
						return
					case "n":
						if err := stacks[s].Change("rm", run.changeName); err != nil {
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
				run.stacks = make(map[string]string)
				// stack list
				stklist := make([]string, len(stacks))
				i := 0
				for k := range stacks {
					stklist[i] = k
					i++
				}

				// create checklist
				choices := c.Checklist(
					stklist,
					fmt.Sprintf("select stacks to %s:", log.ColorString("set-policy", "yellow")),
					nil,
				)

				// define run.stacks
				run.stacks = make(map[string]string)
				for _, i := range choices {
					if i < 0 {
						fmt.Printf("--\nPress %s to return\n--\n", log.ColorString("ENTER", "green"))
						return
					}
					run.stacks[stklist[i]] = ""
				}

				for s := range run.stacks {
					wg.Add(1)
					go func(s string) {
						if _, ok := stacks[s]; !ok {
							log.Error("stack [%s] not found in config", s)
							wg.Done()
							return
						}

						if err := stacks[s].StackPolicy(); err != nil {
							log.Error(err.Error())
						}

						wg.Done()
						return

					}(s)
				}

				wg.Wait()

			},
		},

		// reload command
		&ishell.Cmd{
			Name:     "reload",
			Help:     "reload Qaz configuration source into shell environment",
			LongHelp: "reload",
			Func: func(c *ishell.Context) {
				log.Debug("%v", stacks)
				// off load stacks
				for k := range stacks {
					delete(stacks, k)
				}

				log.Debug("%v", stacks)

				// re-read config
				err := Configure(run.cfgSource, run.cfgRaw)
				utils.HandleError(err)
				log.Info("config reloaded: [%s]", run.cfgSource)
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
