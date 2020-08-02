package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	yaml "gopkg.in/yaml.v2"

	"github.com/daidokoro/qaz/stacks"
	"github.com/daidokoro/qaz/utils"

	"github.com/CrowdSurge/banner"
	"github.com/daidokoro/qaz/log"
	"github.com/spf13/cobra"
)

var t sync.Map

// initialise - adds, logging and repo vars to dependecny functions
var initialise = func(cmd *cobra.Command, args []string) {
	// add logging
	log.SetDefault(log.NewDefaultLogger(run.debug, run.colors))
	log.Debug("initialising command [%s]", cmd.Name())

	// add repo
	stacks.Git(&gitrepo)
}

var (
	// RootCmd command (calls all other commands)
	RootCmd = &cobra.Command{
		Use:   "qaz",
		Short: version,
		Run: func(cmd *cobra.Command, args []string) {

			if run.version {
				fmt.Printf("qaz - Version %s"+"\n", version)
				return
			}

			cmd.Help()
		},
	}

	// initCmd used to initial project
	initCmd = &cobra.Command{
		Use:    "init [target directory]",
		Short:  "Creates an initial Qaz config file",
		PreRun: initialise,
		Run: func(cmd *cobra.Command, args []string) {

			// Print Banner
			banner.Print("qaz")
			fmt.Printf("\n--\n")

			var target string
			switch len(args) {
			case 0:
				target, _ = os.Getwd()
			default:
				target = args[0]
			}

			// Get Project & AWS Region
			arrow := log.ColorString("->", log.MAGENTA)
			project = utils.GetInput(fmt.Sprintf("%s Enter your Project name", arrow), "qaz-project")
			region := utils.GetInput(fmt.Sprintf("%s Enter AWS Region", arrow), "eu-west-1")

			// set target paths
			c := filepath.Join(target, "config.yml")

			// Check if config file exists
			var overwrite string
			if _, err := os.Stat(c); err == nil {
				overwrite = utils.GetInput(
					fmt.Sprintf(
						"%s [%s] already exist, Do you want to %s?(Y/N) ",
						log.ColorString("->", log.YELLOW),
						c, log.ColorString("Overwrite", log.RED)),
					"N",
				)

				if overwrite == "Y" {
					fmt.Println(fmt.Sprintf("%s Overwriting: [%s]..", log.ColorString("->", log.YELLOW), c))
				}
			}

			// Create template file
			if overwrite != "N" {
				if err := ioutil.WriteFile(c, utils.ConfigTemplate(project, region), 0644); err != nil {
					fmt.Printf("%s Error, unable to create config.yml file: %s"+"\n", err, log.ColorString("->", log.RED))
					return
				}
			}

			fmt.Println("--")
		},
	}

	// set stack policy
	policyCmd = &cobra.Command{
		Use:     "set-policy",
		Short:   "Set Stack Policies based on configured value",
		Example: "qaz set-policy <stack name>",
		PreRun:  initialise,
		Run: func(cmd *cobra.Command, args []string) {
			var wg sync.WaitGroup
			if len(args) == 0 {
				utils.HandleError(fmt.Errorf("please specify stack name"))
			}

			stks, err := Configure(run.cfgSource, run.cfgRaw)
			utils.HandleError(err)

			for _, s := range args {
				wg.Add(1)
				go func(s string) {
					if _, ok := stks.Get(s); !ok {
						utils.HandleError(fmt.Errorf("Stack [%s] not found in config", s))
					}

					err := stks.MustGet(s).StackPolicy()
					utils.HandleError(err)

					wg.Done()
					return

				}(s)
			}

			wg.Wait()

		},
	}

	// Values - print json config values for a stack
	valuesCmd = &cobra.Command{
		Use:     "values [stack]",
		Short:   "Print stack values from config in YAML format",
		Example: "qaz values stack",
		PreRun:  initialise,
		Run: func(cmd *cobra.Command, args []string) {

			if len(args) == 0 {
				utils.HandleError(fmt.Errorf("please specify stack name"))
				return
			}

			// set stack value based on argument
			s := args[0]

			stks, err := Configure(run.cfgSource, run.cfgRaw)
			utils.HandleError(err)

			if _, ok := stks.Get(s); !ok {
				utils.HandleError(fmt.Errorf("Stack [%s] not found in config", s))
			}

			values := stks.MustGet(s).TemplateValues[s].(map[string]interface{})

			log.Debug("Converting stack outputs to JSON from: %s", values)
			output, err := yaml.Marshal(values)
			utils.HandleError(err)

			reg, err := regexp.Compile(stacks.OutputRegex)
			utils.HandleError(err)

			resp := reg.ReplaceAllStringFunc(string(output), func(s string) string {
				return log.ColorString(s, log.CYAN)
			})

			fmt.Printf("\n%s\n", resp)
		},
	}
)
