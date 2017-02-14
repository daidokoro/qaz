package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/CrowdSurge/banner"
	"github.com/spf13/cobra"
)

// job var used as a central point for command data
var job = struct {
	cfgFile      string
	tplFile      string
	profile      string
	tplFiles     []string
	stacks       map[string]string
	terminateAll bool
	version      bool
	request      string
	debug        bool
}{}

// Wait Group for handling goroutines
var wg sync.WaitGroup

// root command (calls all other commands)
var rootCmd = &cobra.Command{
	Use:   "qaze",
	Short: fmt.Sprintf("%s\n--> Shut up & deploy my templates...!", colorString(banner.PrintS("qaze"), "magenta")),
	Run: func(cmd *cobra.Command, args []string) {

		if job.version {
			fmt.Printf("qaze - Version %s"+"\n", version)
			return
		}

		cmd.Help()
	},
}

var initCmd = &cobra.Command{
	Use:   "init [target directory]",
	Short: "Creates a basic qaze project",
	Run: func(cmd *cobra.Command, args []string) {

		// Print Banner
		banner.Print("qaze")
		fmt.Printf("\n--\n")

		var target string
		switch len(args) {
		case 0:
			target, _ = os.Getwd()
		default:
			target = args[0]
		}

		// Get Project & AWS Region
		project = getInput("-> Enter your Project name", "MyqazeProject")
		region = getInput("-> Enter AWS Region", "eu-west-1")

		// set target paths
		c := filepath.Join(target, "config.yml")
		t := filepath.Join(target, "templates")
		f := filepath.Join(target, "files")

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

		// Create template folder
		for _, dir := range []string{t, f} {
			if err := os.Mkdir(dir, os.ModePerm); err != nil {
				fmt.Printf("%s [%s] folder not created: %s"+"\n--\n", colorString("->", "yellow"), dir, err)
				return
			}
		}

		fmt.Println("--")
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates a JSON or YAML template",
	Run: func(cmd *cobra.Command, args []string) {

		job.request = "generate"

		s, source, err := getSource(job.tplFile)
		if err != nil {
			handleError(err)
			return
		}

		job.tplFile = source
		err = configReader(job.cfgFile)
		if err != nil {
			handleError(err)
			return
		}

		name := fmt.Sprintf("%s-%s", project, s)
		Log(fmt.Sprintln("Generating a template for ", name), "debug")

		tpl, err := templateParser(job.tplFile)
		if err != nil {
			handleError(err)
			return
		}
		fmt.Println(tpl)
	},
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploys stack(s) to AWS",
	Example: strings.Join([]string{
		"qaze deploy -c path/to/config -t path/to/template",
		"qaze deploy -c path/to/config -t stack::s3//bucket/key",
		"qaze deploy -c path/to/config -t stack::http://someurl",
	}, "\n"),
	Run: func(cmd *cobra.Command, args []string) {

		job.request = "deploy"

		job.stacks = make(map[string]string)

		sourceCopy := job.tplFiles

		// creating empty template list for re-population later
		job.tplFiles = []string{}

		for _, src := range sourceCopy {
			if strings.Contains(src, `*`) {
				glob, _ := filepath.Glob(src)

				for _, f := range glob {
					job.tplFiles = append(job.tplFiles, f)
				}
				continue
			}

			job.tplFiles = append(job.tplFiles, src)
		}

		for _, f := range job.tplFiles {
			s, source, err := getSource(f)
			if err != nil {
				handleError(err)
				return
			}
			job.stacks[s] = source
		}

		err := configReader(job.cfgFile)
		if err != nil {
			handleError(err)
			return
		}

		for s, f := range job.stacks {
			if v, err := templateParser(f); err != nil {
				handleError(err)
			} else {
				stacks[s].template = v
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
		"qaze update -c path/to/config -t stack::path/to/template",
		"qaze update -c path/to/config -t stack::s3//bucket/key",
		"qaze update -c path/to/config -t stack::http://someurl",
	}, "\n"),
	Run: func(cmd *cobra.Command, args []string) {

		job.request = "update"

		s, source, err := getSource(job.tplFile)
		if err != nil {
			handleError(err)
			return
		}

		fmt.Println(source)

		job.tplFile = source

		err = configReader(job.cfgFile)
		if err != nil {
			handleError(err)
			return
		}

		v, err := templateParser(job.tplFile)
		if err != nil {
			handleError(err)
			return
		}

		stacks[s].template = v

		// Update stack
		sess, err := awsSession()
		if err != nil {
			handleError(err)
			return
		}
		stacks[s].update(sess)

	},
}

var terminateCmd = &cobra.Command{
	Use:   "terminate [stacks]",
	Short: "Terminates stacks",
	Run: func(cmd *cobra.Command, args []string) {

		job.request = "terminate"

		if !job.terminateAll {
			job.stacks = make(map[string]string)
			for _, stk := range args {
				job.stacks[stk] = ""
			}

			if len(job.stacks) == 0 {
				Log("No stack specified for termination", "warn")
				return
			}
		}

		err := configReader(job.cfgFile)
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

		job.request = "status"

		err := configReader(job.cfgFile)
		if err != nil {
			handleError(err)
			return
		}

		sess, err := awsSession()
		if err != nil {
			handleError(err)
			return
		}

		for _, v := range stacks {
			wg.Add(1)
			go func(s *stack) {
				if err := s.status(sess); err != nil {
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
	Example: "qaze outputs vpc subnets --config path/to/config",
	Run: func(cmd *cobra.Command, args []string) {

		job.request = "outputs"

		if len(args) < 1 {
			fmt.Println("Please specify stack(s) to check, For details try --> qaze outputs --help")
			return
		}

		err := configReader(job.cfgFile)
		if err != nil {
			handleError(err)
			return
		}

		sess, err := awsSession()
		if err != nil {
			handleError(err)
			return
		}

		for _, s := range args {
			wg.Add(1)
			go func(s string) {
				if err := stacks[s].outputs(sess); err != nil {
					handleError(err)
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
	Example: "qaze exports",
	Run: func(cmd *cobra.Command, args []string) {

		job.request = "exports"

		sess, err := awsSession()
		if err != nil {
			handleError(err)
			return
		}

		Exports(sess)

	},
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Validates Cloudformation Templates",
	Example: strings.Join([]string{
		"qaze check -c path/to/config.yml -t path/to/template -c path/to/config",
		"qaze check -c path/to/config.yml -t stack::http://someurl.example",
		"qaze check -c path/to/config.yml -t stack::s3://bucket/key",
	}, "\n"),
	Run: func(cmd *cobra.Command, args []string) {

		job.request = "validate"

		s, source, err := getSource(job.tplFile)
		if err != nil {
			handleError(err)
			return
		}

		job.tplFile = source

		err = configReader(job.cfgFile)
		if err != nil {
			handleError(err)
			return
		}

		name := fmt.Sprintf("%s-%s", project, s)
		fmt.Println("Validating template for", name)

		tpl, err := templateParser(job.tplFile)
		if err != nil {
			handleError(err)
			return
		}

		sess, err := awsSession()
		if err != nil {
			handleError(err)
			return
		}

		if err := Check(tpl, sess); err != nil {
			handleError(err)
			return
		}
	},
}
