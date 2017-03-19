package main

// Init & Logging sit here

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

// Used for mapping log level, may or may not expand in the future..
var level = struct {
	debug string
	warn  string
	err   string
	info  string
}{"debug", "warn", "error", "info"}

// handleError - handleError logs the err and exits the app if err not nil
func handleError(e error) {
	if e != nil {
		log.WithFields(log.Fields{
			"request": job.request,
		}).Errorln(e.Error()) // logrus error calls os.exit after logging
	}
}

// Log - Handles all logging accross app
func Log(msg, lvl string) {

	// TODO: Get this to only check once.
	if job.debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	l := log.WithFields(log.Fields{
		"request": job.request,
	})

	switch lvl {
	case "debug":
		l.Debugln(msg)
	case "warn":
		l.Warnln(msg)
	case "error":
		l.Errorln(msg)
	default:
		l.Infoln(msg)
	}
}

func init() {

	// Define Deploy Flags
	deployCmd.Flags().StringArrayVarP(&job.tplFiles, "template", "t", []string{`./templates/*`}, "path to template file(s) Or stack::url")

	// Define Terminate Flags
	terminateCmd.Flags().BoolVarP(&job.terminateAll, "all", "A", false, "terminate all stacks")

	// Define Output Flags
	outputsCmd.Flags().StringVarP(&job.profile, "profile", "p", "default", "configured aws profile")

	// Define Exports Flags
	exportsCmd.Flags().StringVarP(&region, "region", "r", "eu-west-1", "AWS Region")

	// Define Root Flags
	rootCmd.Flags().BoolVarP(&job.version, "version", "", false, "print current/running version")
	rootCmd.PersistentFlags().StringVarP(&job.profile, "profile", "p", "default", "configured aws profile")
	rootCmd.PersistentFlags().BoolVarP(&job.debug, "debug", "", false, "Run in debug mode...")

	// Define Invoke Flags
	invokeCmd.Flags().StringVarP(&region, "region", "r", "eu-west-1", "AWS Region")
	invokeCmd.Flags().StringVarP(&job.funcEvent, "event", "e", "", "JSON Event data for AWS Lambda invoke")

	// Define Changes Command
	changeCmd.AddCommand(create, rm, list, execute, desc)

	// Add Config --config common flag
	for _, cmd := range []interface{}{tailCmd, checkCmd, updateCmd, outputsCmd, statusCmd, terminateCmd, generateCmd, deployCmd} {
		cmd.(*cobra.Command).Flags().StringVarP(&job.cfgFile, "config", "c", "config.yml", "path to config file")
	}

	// Add Template --template common flag
	for _, cmd := range []interface{}{generateCmd, updateCmd, checkCmd} {
		cmd.(*cobra.Command).Flags().StringVarP(&job.tplFile, "template", "t", "template", "path to template file Or stack::url")
	}

	for _, cmd := range []interface{}{create, list, rm, execute, desc} {
		cmd.(*cobra.Command).Flags().StringVarP(&job.cfgFile, "config", "c", "config.yml", "path to config file [Required]")
		cmd.(*cobra.Command).Flags().StringVarP(&job.stackName, "stack", "s", "", "Qaz local project Stack Name [Required]")
	}

	create.Flags().StringVarP(&job.tplFile, "template", "t", "template", "path to template file Or stack::url")
	changeCmd.Flags().StringVarP(&job.cfgFile, "config", "c", "config.yml", "path to config file")

	rootCmd.AddCommand(
		generateCmd, deployCmd, terminateCmd,
		statusCmd, outputsCmd, initCmd,
		updateCmd, checkCmd, exportsCmd,
		invokeCmd, tailCmd, changeCmd,
	)

	// Setup logging
	log.SetFormatter(new(prefixed.TextFormatter))
}
