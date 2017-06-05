package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"qaz/utils"

	"github.com/CrowdSurge/banner"
	"github.com/spf13/cobra"
)

// RootCmd command (calls all other commands)
var RootCmd = &cobra.Command{
	Use:   "qaz",
	Short: fmt.Sprintf("\n"),
	Run: func(cmd *cobra.Command, args []string) {

		if run.version {
			fmt.Printf("qaz - Version %s"+"\n", version)
			return
		}

		cmd.Help()
	},
}

var initCmd = &cobra.Command{
	Use:   "init [target directory]",
	Short: "Creates an initial Qaz config file",
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

var invokeCmd = &cobra.Command{
	Use:   "invoke",
	Short: "Invoke AWS Lambda Functions",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 1 {
			fmt.Println("No Lambda Function specified")
			return
		}

		sess, err := manager.GetSess(run.profile)
		utils.HandleError(err)

		f := awsLambda{name: args[0]}

		if run.funcEvent != "" {
			f.payload = []byte(run.funcEvent)
		}

		if err := f.Invoke(sess); err != nil {
			if strings.Contains(err.Error(), "Unhandled") {
				utils.HandleError(fmt.Errorf("Unhandled Exception: Potential Issue with Lambda Function Logic for %s...\n", f.name))
			}
			utils.HandleError(err)
		}

		fmt.Println(f.response)

	},
}

var policyCmd = &cobra.Command{
	Use:     "set-policy",
	Short:   "Set Stack Policies based on configured value",
	Example: "qaz set-policy <stack name>",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			utils.HandleError(fmt.Errorf("Please specify stack name..."))
			return
		}

		err := configure(run.cfgSource, run.cfgRaw)
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
