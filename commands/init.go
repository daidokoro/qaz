package commands

// Init sits here

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {

	// Define Deploy Flags
	deployCmd.Flags().StringArrayVarP(&run.tplSources, "template", "t", []string{}, "path to template file(s) Or stack::url")
	deployCmd.Flags().BoolVarP(&run.rollback, "disable-rollback", "", false, "Set Stack to rollback on deployment failures")
	deployCmd.Flags().BoolVarP(&run.all, "all", "A", false, "deploy all stacks with defined Sources in config")

	// Define Git Deploy Flags
	gitDeployCmd.Flags().BoolVarP(&run.rollback, "disable-rollback", "", false, "Set Stack to rollback on deployment failures")
	gitDeployCmd.Flags().BoolVarP(&run.all, "all", "A", false, "deploy all stacks with defined Sources in config")
	gitDeployCmd.Flags().StringVarP(&run.gituser, "user", "u", "", "git username")
	gitDeployCmd.Flags().StringVarP(&run.gitpass, "password", "", "", "git password")
	gitDeployCmd.Flags().StringVarP(&run.gitrsa, "ssh-rsa", "", filepath.Join(os.Getenv("HOME"), ".ssh/id_rsa"), "path to git SSH id_rsa")

	// Define Git Status Command
	gitStatusCmd.Flags().StringVarP(&run.gitrsa, "ssh-rsa", "", filepath.Join(os.Getenv("HOME"), ".ssh/id_rsa"), "path to git SSH id_rsa")
	gitStatusCmd.Flags().StringVarP(&run.gitpass, "password", "", "", "git password")
	gitStatusCmd.Flags().StringVarP(&run.gituser, "user", "u", "", "git username")

	// Define Terminate Flags
	terminateCmd.Flags().BoolVarP(&run.all, "all", "A", false, "terminate all stacks")

	// Define Output Flags
	outputsCmd.Flags().StringVarP(&run.profile, "profile", "p", "default", "configured aws profile")

	// Define Root Flags
	RootCmd.Flags().BoolVarP(&run.version, "version", "", false, "print current/running version")
	RootCmd.PersistentFlags().BoolVarP(&run.colors, "no-colors", "", false, "disable colors in outputs")
	RootCmd.PersistentFlags().StringVarP(&run.profile, "profile", "p", "default", "configured aws profile")
	RootCmd.PersistentFlags().StringVarP(&run.region, "region", "r", "", "configured aws region: if blank, the region is acquired via the profile")
	RootCmd.PersistentFlags().BoolVarP(&run.debug, "debug", "", false, "Run in debug mode...")

	// Define Lambda Invoke Flags
	invokeCmd.Flags().StringVarP(&run.funcEvent, "event", "e", "", "JSON Event data for AWS Lambda invoke")
	invokeCmd.Flags().BoolVarP(&run.lambdAsync, "async", "x", false, "invoke lambda function asynchronously ")

	// Define Changes Command
	changeCmd.AddCommand(create, rm, list, execute, desc)

	// Define Protect Command
	protectCmd.Flags().BoolVarP(&run.protectOff, "off", "", false, "set termination protection to off")
	protectCmd.Flags().BoolVarP(&run.all, "all", "A", false, "protect all stacks")

	// Define Update Command
	updateCmd.Flags().BoolVarP(&run.interactive, "interactive", "i", false, "preview change-set and ask before executing it")

	// Add Config --config common flag
	for _, cmd := range []interface{}{
		checkCmd,
		updateCmd,
		outputsCmd,
		statusCmd,
		terminateCmd,
		generateCmd,
		deployCmd,
		gitDeployCmd,
		gitStatusCmd,
		policyCmd,
		valuesCmd,
		shellCmd,
		protectCmd,
		lintCmd,
	} {
		cmd.(*cobra.Command).Flags().StringVarP(&run.cfgSource, "config", "c", defaultConfig(), "path to config file")
	}

	// Add Template --template common flag
	for _, cmd := range []interface{}{
		generateCmd,
		updateCmd,
		checkCmd,
		lintCmd,
	} {
		cmd.(*cobra.Command).Flags().StringVarP(&run.tplSource, "template", "t", "", "path to template source Or stack::source")
	}

	for _, cmd := range []interface{}{
		create,
		list,
		rm,
		execute,
		desc,
	} {
		cmd.(*cobra.Command).Flags().StringVarP(&run.cfgSource, "config", "c", defaultConfig(), "path to config file (Required)")
		cmd.(*cobra.Command).Flags().StringVarP(&run.stackName, "stack", "s", "", "Qaz local project Stack Name (Required)")
	}

	create.Flags().StringVarP(&run.tplSource, "template", "t", "", "path to template file Or stack::url")
	changeCmd.Flags().StringVarP(&run.cfgSource, "config", "c", defaultConfig(), "path to config file")

	// add commands
	RootCmd.AddCommand(
		generateCmd,
		deployCmd,
		terminateCmd,
		statusCmd,
		outputsCmd,
		initCmd,
		updateCmd,
		checkCmd,
		exportsCmd,
		invokeCmd,
		changeCmd,
		policyCmd,
		gitDeployCmd,
		gitStatusCmd,
		valuesCmd,
		shellCmd,
		protectCmd,
		completionCmd,
		lintCmd,
	)

}
