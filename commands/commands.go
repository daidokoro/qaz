package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	yaml "gopkg.in/yaml.v2"

	"github.com/daidokoro/qaz/bucket"
	"github.com/daidokoro/qaz/repo"
	stks "github.com/daidokoro/qaz/stacks"
	"github.com/daidokoro/qaz/utils"

	"github.com/CrowdSurge/banner"
	"github.com/spf13/cobra"
)

// initialise - adds, logging and repo vars to dependecny functions
var initialise = func(cmd *cobra.Command, args []string) {
	log.Debug(fmt.Sprintf("Initialising Command [%s]", cmd.Name()))
	// add logging
	stks.Log = &log
	bucket.Log = &log
	repo.Log = &log
	utils.Log = &log

	// add repo
	stks.Git = &gitrepo
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
			arrow := log.ColorString("->", "magenta")
			project = utils.GetInput(fmt.Sprintf("%s Enter your Project name", arrow), "qaz-project")
			region = utils.GetInput(fmt.Sprintf("%s Enter AWS Region", arrow), "eu-west-1")

			// set target paths
			c := filepath.Join(target, "config.yml")

			// Check if config file exists
			var overwrite string
			if _, err := os.Stat(c); err == nil {
				overwrite = utils.GetInput(
					fmt.Sprintf("%s [%s] already exist, Do you want to %s?(Y/N) ", log.ColorString("->", "yellow"), c, log.ColorString("Overwrite", "red")),
					"N",
				)

				if overwrite == "Y" {
					fmt.Println(fmt.Sprintf("%s Overwriting: [%s]..", log.ColorString("->", "yellow"), c))
				}
			}

			// Create template file
			if overwrite != "N" {
				if err := ioutil.WriteFile(c, utils.ConfigTemplate(project, region), 0644); err != nil {
					fmt.Printf("%s Error, unable to create config.yml file: %s"+"\n", err, log.ColorString("->", "red"))
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

			if len(args) == 0 {
				utils.HandleError(fmt.Errorf("Please specify stack name..."))
			}

			err := Configure(run.cfgSource, run.cfgRaw)
			utils.HandleError(err)

			for _, s := range args {
				wg.Add(1)
				go func(s string) {
					if _, ok := stacks[s]; !ok {
						utils.HandleError(fmt.Errorf("Stack [%s] not found in config", s))
					}

					err := stacks[s].StackPolicy()
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
				utils.HandleError(fmt.Errorf("Please specify stack name..."))
				return
			}

			// set stack value based on argument
			s := args[0]

			err := Configure(run.cfgSource, run.cfgRaw)
			utils.HandleError(err)

			if _, ok := stacks[s]; !ok {
				utils.HandleError(fmt.Errorf("Stack [%s] not found in config", s))
			}

			values := stacks[s].TemplateValues[s].(map[string]interface{})

			log.Debug(fmt.Sprintln("Converting stack outputs to JSON from:", values))
			output, err := yaml.Marshal(values)
			utils.HandleError(err)

			reg, err := regexp.Compile(".+?:(\n| )")
			utils.HandleError(err)

			resp := reg.ReplaceAllStringFunc(string(output), func(s string) string {
				return log.ColorString(s, "cyan")
			})

			fmt.Printf("\n%s\n", resp)
		},
	}
)
