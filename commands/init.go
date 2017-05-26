package commands

// Init sits here

import "github.com/spf13/cobra"

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

	// Define Terminate Flags
	terminateCmd.Flags().BoolVarP(&run.all, "all", "A", false, "terminate all stacks")

	// Define Output Flags
	outputsCmd.Flags().StringVarP(&run.profile, "profile", "p", "default", "configured aws profile")

	// Define Exports Flags
	exportsCmd.Flags().StringVarP(&region, "region", "r", "eu-west-1", "AWS Region")

	// Define Root Flags
	RootCmd.Flags().BoolVarP(&run.version, "version", "", false, "print current/running version")
	RootCmd.PersistentFlags().BoolVarP(&run.colors, "no-colors", "", false, "disable colors in outputs")
	RootCmd.PersistentFlags().StringVarP(&run.profile, "profile", "p", "default", "configured aws profile")
	RootCmd.PersistentFlags().BoolVarP(&run.debug, "debug", "", false, "Run in debug mode...")

	// Define Invoke Flags
	invokeCmd.Flags().StringVarP(&region, "region", "r", "eu-west-1", "AWS Region")
	invokeCmd.Flags().StringVarP(&run.funcEvent, "event", "e", "", "JSON Event data for AWS Lambda invoke")

	// Define Changes Command
	changeCmd.AddCommand(create, rm, list, execute, desc)

	// Add Config --config common flag
	for _, cmd := range []interface{}{checkCmd, updateCmd, outputsCmd, statusCmd, terminateCmd, generateCmd, deployCmd, gitDeployCmd, policyCmd} {
		cmd.(*cobra.Command).Flags().StringVarP(&run.cfgSource, "config", "c", defaultConfig(), "path to config file")
	}

	// Add Template --template common flag
	for _, cmd := range []interface{}{generateCmd, updateCmd, checkCmd} {
		cmd.(*cobra.Command).Flags().StringVarP(&run.tplSource, "template", "t", "", "path to template source Or stack::source")
	}

	for _, cmd := range []interface{}{create, list, rm, execute, desc} {
		cmd.(*cobra.Command).Flags().StringVarP(&run.cfgSource, "config", "c", defaultConfig(), "path to config file [Required]")
		cmd.(*cobra.Command).Flags().StringVarP(&run.stackName, "stack", "s", "", "Qaz local project Stack Name [Required]")
	}

	create.Flags().StringVarP(&run.tplSource, "template", "t", "", "path to template file Or stack::url")
	changeCmd.Flags().StringVarP(&run.cfgSource, "config", "c", defaultConfig(), "path to config file")

	RootCmd.AddCommand(
		generateCmd, deployCmd, terminateCmd,
		statusCmd, outputsCmd, initCmd,
		updateCmd, checkCmd, exportsCmd,
		invokeCmd, changeCmd, policyCmd,
		gitDeployCmd,
	)

}
