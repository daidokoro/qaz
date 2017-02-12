package main

import (
	"fmt"
	"io/ioutil"
	"log"
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
}{}

// Wait Group for handling goroutines
var wg sync.WaitGroup

func init() {
	// Define Generate Flags
	generateCmd.Flags().StringVarP(&job.cfgFile, "config", "c", "config.yml", "path to config file")
	generateCmd.Flags().StringVarP(&job.tplFile, "template", "t", "template", "path to template file Or stack::url")

	// Define Deploy Flags
	deployCmd.Flags().StringVarP(&job.cfgFile, "config", "c", "config.yml", "path to config file")
	deployCmd.Flags().StringArrayVarP(&job.tplFiles, "template", "t", []string{`./templates/*`}, "path to template file(s) Or stack::url")

	// Define Terminate Flags
	terminateCmd.Flags().StringVarP(&job.cfgFile, "config", "c", "config.yml", "path to config file")
	terminateCmd.Flags().BoolVarP(&job.terminateAll, "all", "A", false, "terminate all stacks")

	// Define Status Flags
	statusCmd.Flags().StringVarP(&job.cfgFile, "config", "c", "config.yml", "path to config file")

	// Define Output Flags
	outputsCmd.Flags().StringVarP(&job.cfgFile, "config", "c", "config.yml", "path. to config file")
	outputsCmd.Flags().StringVarP(&job.profile, "profile", "p", "default", "configured aws profile")

	// Define Update Flags
	updateCmd.Flags().StringVarP(&job.cfgFile, "config", "c", "config.yml", "path to config file")
	updateCmd.Flags().StringVarP(&job.tplFile, "template", "t", "", "path to template file Or stack::url [Required]")

	// Define Check Flags
	checkCmd.Flags().StringVarP(&job.cfgFile, "config", "c", "config.yml", "path to config file")
	checkCmd.Flags().StringVarP(&job.tplFile, "template", "t", "template", "path to template file Or stack::url")

	// Define Exports Flags
	exportsCmd.Flags().StringVarP(&region, "region", "r", "eu-west-1", "AWS Region")

	// Define Root Flags
	rootCmd.Flags().BoolVarP(&job.version, "version", "", false, "print current/running version")
	rootCmd.PersistentFlags().StringVarP(&job.profile, "profile", "p", "default", "configured aws profile")
	rootCmd.AddCommand(generateCmd, deployCmd, terminateCmd, statusCmd, outputsCmd, initCmd, updateCmd, checkCmd, exportsCmd)
}

// root command (calls all other commands)
var rootCmd = &cobra.Command{
	Use:   "qaze",
	Short: "qaze is a simple wrapper around Cloudformation. ",
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

		s, source, err := getSource(job.tplFile)
		if err != nil {
			log.Println(err.Error())
			return
		}

		job.tplFile = source
		err = configReader(job.cfgFile)
		if err != nil {
			log.Println(err.Error())
			return
		}

		name := fmt.Sprintf("%s-%s", project, s)
		log.Println("Generating a template for ", name)

		tpl, err := templateParser(job.tplFile)

		if err != nil {
			log.Println(err.Error())
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
				log.Println(err.Error())
				return
			}
			job.stacks[s] = source
		}

		err := configReader(job.cfgFile)
		if err != nil {
			log.Println(err.Error())
			return
		}

		for s, f := range job.stacks {
			if v, err := templateParser(f); err != nil {
				log.Println("Error", err.Error())
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

		s, source, err := getSource(job.tplFile)
		if err != nil {
			log.Println(err.Error())
			return
		}

		fmt.Println(source)

		job.tplFile = source

		err = configReader(job.cfgFile)
		if err != nil {
			log.Println(err.Error())
			return
		}

		v, err := templateParser(job.tplFile)
		if err != nil {
			log.Println("Error", err.Error())
			return
		}

		stacks[s].template = v

		// Update stack
		sess, err := awsSession()
		if err != nil {
			fmt.Println("Error:", err.Error())
			return
		}
		stacks[s].update(sess)

	},
}

var terminateCmd = &cobra.Command{
	Use:   "terminate [stacks]",
	Short: "Terminates stacks",
	Run: func(cmd *cobra.Command, args []string) {

		if !job.terminateAll {
			job.stacks = make(map[string]string)
			for _, stk := range args {
				job.stacks[stk] = ""
			}

			if len(job.stacks) == 0 {
				log.Println("No stack specified for termination")
				return
			}
		}

		err := configReader(job.cfgFile)
		if err != nil {
			log.Println(err.Error())
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

		err := configReader(job.cfgFile)
		if err != nil {
			log.Println(err.Error())
			return
		}

		sess, err := awsSession()
		if err != nil {
			log.Println("Error: ", err.Error())
			return
		}

		for _, v := range stacks {
			wg.Add(1)
			go func(s *stack) {
				s.status(sess)
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
		if len(args) < 1 {
			fmt.Println("Please specify stack(s) to check, For details try --> qaze outputs --help")
		}

		err := configReader(job.cfgFile)
		if err != nil {
			log.Println(err.Error())
			return
		}

		sess, err := awsSession()
		if err != nil {
			log.Println("Error: ", err.Error())
			return
		}

		for _, s := range args {
			wg.Add(1)
			go func(s string) {
				stacks[s].outputs(sess)
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

		sess, err := awsSession()
		if err != nil {
			log.Println("Error: ", err.Error())
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

		s, source, err := getSource(job.tplFile)
		if err != nil {
			log.Println(err.Error())
			return
		}

		job.tplFile = source

		err = configReader(job.cfgFile)
		if err != nil {
			log.Println(err.Error())
			return
		}

		name := fmt.Sprintf("%s-%s", project, s)
		log.Println("Validating template for ", name)

		tpl, err := templateParser(job.tplFile)
		if err != nil {
			log.Println(err.Error())
			return
		}

		sess, err := awsSession()
		if err != nil {
			log.Println("Error: ", err.Error())
			return
		}

		if err := Check(tpl, sess); err != nil {
			fmt.Println(err.Error())
			return
		}
	},
}
