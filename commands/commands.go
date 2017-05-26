package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/CrowdSurge/banner"
	"github.com/spf13/cobra"
)

// config environment variable
const configENV = "QAZ_CONFIG"

// run.var used as a central point for command data
var run = struct {
	cfgSource  string
	tplSource  string
	profile    string
	tplSources []string
	stacks     map[string]string
	all        bool
	version    bool
	request    string
	debug      bool
	funcEvent  string
	changeName string
	stackName  string
	rollback   bool
	colors     bool
	cfgRaw     string
	gituser    string
	gitpass    string
}{}

// Wait Group for handling goroutines
var wg sync.WaitGroup

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
		arrow := colorString("->", "magenta")
		project = getInput(fmt.Sprintf("%s Enter your Project name", arrow), "qaz-project")
		region = getInput(fmt.Sprintf("%s Enter AWS Region", arrow), "eu-west-1")

		// set target paths
		c := filepath.Join(target, "config.yml")

		// Check if config file exists
		var overwrite string
		if _, err := os.Stat(c); err == nil {
			overwrite = getInput(
				fmt.Sprintf("%s [%s] already exist, Do you want to %s?(Y/N) ", colorString("->", "yellow"), c, colorString("Overwrite", "red")),
				"N",
			)

			if overwrite == "Y" {
				fmt.Println(fmt.Sprintf("%s Overwriting: [%s]..", colorString("->", "yellow"), c))
			}
		}

		// Create template file
		if overwrite != "N" {
			if err := ioutil.WriteFile(c, configTemplate(project, region), 0644); err != nil {
				fmt.Printf("%s Error, unable to create config.yml file: %s"+"\n", err, colorString("->", "red"))
				return
			}
		}

		fmt.Println("--")
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate [stack]",
	Short: "Generates template from configuration values",
	Example: strings.Join([]string{
		"",
		"qaz generate -c config.yml -t stack::source",
		"qaz generate vpc -c config.yml",
	}, "\n"),
	Run: func(cmd *cobra.Command, args []string) {

		var s string
		var source string

		err := configReader(run.cfgSource, run.cfgRaw)
		if err != nil {
			handleError(err)
			return
		}

		if run.tplSource != "" {
			s, source, err = getSource(run.tplSource)
			if err != nil {
				handleError(err)
				return
			}
		}

		if len(args) > 0 {
			s = args[0]
		}

		// check if stack exists in config
		if _, ok := stacks[s]; !ok {
			handleError(fmt.Errorf("Stack [%s] not found in config", s))
			return
		}

		if stacks[s].source == "" {
			stacks[s].source = source
		}

		name := fmt.Sprintf("%s-%s", project, s)
		Log(fmt.Sprintln("Generating a template for ", name), "debug")

		err = stacks[s].genTimeParser()
		if err != nil {
			handleError(err)
			return
		}
		fmt.Println(stacks[s].template)
	},
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploys stack(s) to AWS",
	Example: strings.Join([]string{
		"qaz deploy stack -c path/to/config",
		"qaz deploy -c path/to/config -t stack::s3://bucket/key",
		"qaz deploy -c path/to/config -t stack::path/to/template",
		"qaz deploy -c path/to/config -t stack::http://someurl",
		"qaz deploy -c path/to/config -t stack::lambda:{some:json}@lambda_function",
	}, "\n"),
	Run: func(cmd *cobra.Command, args []string) {

		err := configReader(run.cfgSource, run.cfgRaw)
		if err != nil {
			handleError(err)
			return
		}

		run.stacks = make(map[string]string)

		// Add run.stacks based on templates Flags
		for _, src := range run.tplSources {
			s, source, err := getSource(src)
			if err != nil {
				handleError(err)
				return
			}
			run.stacks[s] = source
		}

		// Add all stacks with defined sources if all
		if run.all {
			for s, v := range stacks {
				// so flag values aren't overwritten
				if _, ok := run.stacks[s]; !ok {
					run.stacks[s] = v.source
				}
			}
		}

		// Add run.stacks based on Args
		if len(args) > 0 && !run.all {
			for _, stk := range args {
				if _, ok := stacks[stk]; !ok {
					handleError(fmt.Errorf("Stack [%s] not found in conig", stk))
					return
				}
				run.stacks[stk] = stacks[stk].source
			}
		}

		for s, src := range run.stacks {
			if stacks[s].source == "" {
				stacks[s].source = src
			}
			if err := stacks[s].genTimeParser(); err != nil {
				handleError(err)
			} else {

				// Handle missing stacks
				if stacks[s] == nil {
					handleError(fmt.Errorf("Missing Stack in %s: [%s]", run.cfgSource, s))
					return
				}
			}
		}

		// Deploy Stacks
		DeployHandler()

	},
}

var gitDeployCmd = &cobra.Command{
	Use:   "git-deploy",
	Short: "Deploy project from Git repository",
	Run: func(cmd *cobra.Command, args []string) {
		repo, err := NewRepo(args[0])
		if err != nil {
			handleError(err)
			return
		}

		// Passing repo to the global var
		gitrepo = *repo

		if _, ok := repo.files[run.cfgSource]; !ok {
			handleError(fmt.Errorf("Config [%s] not found in git repo", run.cfgSource))
			return
		}

		if err := configReader(run.cfgSource, repo.files[run.cfgSource]); err != nil {
			handleError(err)
			return
		}

		//create run stacks
		run.stacks = make(map[string]string)

		for s, v := range stacks {
			// populate run stacks
			run.stacks[s] = v.source
			if err := stacks[s].genTimeParser(); err != nil {
				handleError(err)
			}
		}

		// Deploy Stacks
		DeployHandler()

	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates a given stack",
	Example: strings.Join([]string{
		"qaz update -c path/to/config -t stack::path/to/template",
		"qaz update -c path/to/config -t stack::s3://bucket/key",
		"qaz update -c path/to/config -t stack::http://someurl",
		"qaz deploy -c path/to/config -t stack::lambda:{some:json}@lambda_function",
	}, "\n"),
	Run: func(cmd *cobra.Command, args []string) {

		var s string
		var source string

		err := configReader(run.cfgSource, run.cfgRaw)
		if err != nil {
			handleError(err)
			return
		}

		if run.tplSource != "" {
			s, source, err = getSource(run.tplSource)
			if err != nil {
				handleError(err)
				return
			}
		}

		if len(args) > 0 {
			s = args[0]
		}

		// check if stack exists in config
		if _, ok := stacks[s]; !ok {
			handleError(fmt.Errorf("Stack [%s] not found in config", s))
			return
		}

		if stacks[s].source == "" {
			stacks[s].source = source
		}

		err = stacks[s].genTimeParser()
		if err != nil {
			handleError(err)
			return
		}

		// Handle missing stacks
		if stacks[s] == nil {
			handleError(fmt.Errorf("Missing Stack in %s: [%s]", run.cfgSource, s))
			return
		}

		if err := stacks[s].update(); err != nil {
			handleError(err)
			return
		}

	},
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Validates Cloudformation Templates",
	Example: strings.Join([]string{
		"qaz check -c path/to/config.yml -t path/to/template -c path/to/config",
		"qaz check -c path/to/config.yml -t stack::http://someurl",
		"qaz check -c path/to/config.yml -t stack::s3://bucket/key",
		"qaz deploy -c path/to/config.yml -t stack::lambda:{some:json}@lambda_function",
	}, "\n"),
	Run: func(cmd *cobra.Command, args []string) {

		var s string
		var source string

		err := configReader(run.cfgSource, "")
		if err != nil {
			handleError(err)
			return
		}

		if run.tplSource != "" {
			s, source, err = getSource(run.tplSource)
			if err != nil {
				handleError(err)
				return
			}
		}

		if len(args) > 0 {
			s = args[0]
		}

		// check if stack exists in config
		if _, ok := stacks[s]; !ok {
			handleError(fmt.Errorf("Stack [%s] not found in config", s))
			return
		}

		if stacks[s].source == "" {
			stacks[s].source = source
		}

		name := fmt.Sprintf("%s-%s", config.Project, s)
		fmt.Println("Validating template for", name)

		if err = stacks[s].genTimeParser(); err != nil {
			handleError(err)
			return
		}

		if err := stacks[s].check(); err != nil {
			handleError(err)
			return
		}
	},
}

var terminateCmd = &cobra.Command{
	Use:   "terminate [stacks]",
	Short: "Terminates stacks",
	Run: func(cmd *cobra.Command, args []string) {

		if !run.all {
			run.stacks = make(map[string]string)
			for _, stk := range args {
				run.stacks[stk] = ""
			}

			if len(run.stacks) == 0 {
				Log("No stack specified for termination", level.warn)
				return
			}
		}

		err := configReader(run.cfgSource, "")
		if err != nil {
			handleError(err)
			return
		}

		// Terminate Stacks
		TerminateHandler()
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Prints status of deployed/un-deployed stacks",
	Run: func(cmd *cobra.Command, args []string) {

		err := configReader(run.cfgSource, run.cfgRaw)
		if err != nil {
			handleError(err)
			return
		}

		for _, v := range stacks {
			wg.Add(1)
			go func(s *stack) {
				if err := s.status(); err != nil {
					handleError(err)
				}
				wg.Done()
			}(v)

		}
		wg.Wait()
	},
}

var outputsCmd = &cobra.Command{
	Use:     "outputs [stack]",
	Short:   "Prints stack outputs",
	Example: "qaz outputs vpc subnets --config path/to/config",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 1 {
			fmt.Println("Please specify stack(s) to check, For details try --> qaz outputs --help")
			return
		}

		err := configReader(run.cfgSource, run.cfgRaw)
		if err != nil {
			handleError(err)
			return
		}

		for _, s := range args {
			// check if stack exists
			if _, ok := stacks[s]; !ok {
				handleError(fmt.Errorf("%s: does not Exist in Config", s))
				continue
			}

			wg.Add(1)
			go func(s string) {
				if err := stacks[s].outputs(); err != nil {
					handleError(err)
					wg.Done()
					return
				}

				for _, i := range stacks[s].output.Stacks {
					m, err := json.MarshalIndent(i.Outputs, "", "  ")
					if err != nil {
						handleError(err)

					}

					fmt.Println(string(m))
				}

				wg.Done()
			}(s)
		}
		wg.Wait()

	},
}

var exportsCmd = &cobra.Command{
	Use:     "exports",
	Short:   "Prints stack exports",
	Example: "qaz exports",
	Run: func(cmd *cobra.Command, args []string) {

		sess, err := manager.GetSess(run.profile)
		if err != nil {
			handleError(err)
			return
		}

		Exports(sess)

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
		if err != nil {
			handleError(err)
			return
		}

		f := awsLambda{name: args[0]}

		if run.funcEvent != "" {
			f.payload = []byte(run.funcEvent)
		}

		if err := f.Invoke(sess); err != nil {
			if strings.Contains(err.Error(), "Unhandled") {
				handleError(fmt.Errorf("Unhandled Exception: Potential Issue with Lambda Function Logic for %s...\n", f.name))
			}
			handleError(err)
			return
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
			handleError(fmt.Errorf("Please specify stack name..."))
			return
		}

		err := configReader(run.cfgSource, run.cfgRaw)
		if err != nil {
			handleError(err)
			return
		}

		for _, s := range args {
			wg.Add(1)
			go func(s string) {

				if _, ok := stacks[s]; !ok {
					handleError(fmt.Errorf("Stack [%s] not found in config", s))

				} else {
					if err := stacks[s].stackPolicy(); err != nil {
						handleError(err)
					}
				}

				wg.Done()
				return

			}(s)
		}

		wg.Wait()

	},
}
