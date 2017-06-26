package commands

import (
	"encoding/json"
	"fmt"
	"qaz/utils"
	"regexp"

	yaml "gopkg.in/yaml.v2"

	"github.com/abiosoft/ishell"
	stks "github.com/daidokoro/qaz/stacks"
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
							log.Error(fmt.Sprintf("failed to fetch status for [%s]: %s", s.Stackname, err.Error()))
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
			Name: "outputs",
			Help: "Prints stack outputs",
			Func: func(c *ishell.Context) {
				if len(c.Args) < 1 {
					fmt.Println("Please specify stack(s) to check, For details try --> qaz outputs --help")
					return
				}

				for _, s := range c.Args {
					// check if stack exists
					if _, ok := stacks[s]; !ok {
						log.Error(fmt.Sprintf("%s: does not Exist in Config", s))
						return
					}

					wg.Add(1)
					go func(s string) {
						if err := stacks[s].Outputs(); err != nil {
							log.Error(err.Error())
							wg.Done()
							return
						}

						for _, i := range stacks[s].Output.Stacks {
							m, err := json.MarshalIndent(i.Outputs, "", "  ")
							if err != nil {
								log.Error(err.Error())
							}
							fmt.Println(string(m))
						}

						wg.Done()
					}(s)
				}
				wg.Wait()
			},
		},

		// values command
		&ishell.Cmd{
			Name: "values",
			Help: "print stack values from config in YAML format",
			Func: func(c *ishell.Context) {
				if len(c.Args) == 0 {
					utils.HandleError(fmt.Errorf("Please specify stack name..."))
					return
				}

				// set stack value based on argument
				s := c.Args[0]

				if _, ok := stacks[s]; !ok {
					log.Error(fmt.Sprintf("Stack [%s] not found in config", s))
				}

				values := stacks[s].TemplateValues[s].(map[string]interface{})

				log.Debug(fmt.Sprintln("Converting stack outputs to JSON from:", values))
				output, err := yaml.Marshal(values)
				if err != nil {
					log.Error(err.Error())
				}

				reg, err := regexp.Compile(".+?:(\n| )")
				if err != nil {
					log.Error(err.Error())
				}

				resp := reg.ReplaceAllStringFunc(string(output), func(s string) string {
					return log.ColorString(s, "cyan")
				})

				fmt.Printf("\n%s\n", resp)
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
