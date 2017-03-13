package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var changeCmd = &cobra.Command{
	Use:   "change",
	Short: "Change-Set management for AWS Stacks",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()

	},
}

var create = &cobra.Command{
	Use:   "create",
	Short: "Create Changet-Set",
	Run: func(cmd *cobra.Command, args []string) {
		job.request = "change-set create"

		if len(args) < 1 {
			fmt.Println("Please provide Change-Set Name...")
			return
		}

		job.changeName = args[0]

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

		v, err := genTimeParser(job.tplFile)
		if err != nil {
			handleError(err)
			return
		}

		// Handle missing stacks
		if stacks[s] == nil {
			handleError(fmt.Errorf("Missing Stack in %s: [%s]", job.cfgFile, s))
			return
		}

		stacks[s].template = v

		// resolve deploy time function
		if err = stacks[s].deployTimeParser(); err != nil {
			handleError(err)
		}

		// create session
		sess, err := awsSession()
		if err != nil {
			handleError(err)
			return
		}

		if err := stacks[s].change(sess, "create"); err != nil {
			handleError(err)
		}

	},
}

var rm = &cobra.Command{
	Use:   "rm",
	Short: "Delete Change-Set",
	Run: func(cmd *cobra.Command, args []string) {
		job.request = "change-set delete"

		if len(args) < 1 {
			fmt.Println("Please provide Change-Set Name...")
			return
		}

		if job.stackName == "" {
			fmt.Println("Please specify stack name using --stack OR -s ...")
			return
		}

		job.changeName = args[0]

		err := configReader(job.cfgFile)
		if err != nil {
			handleError(err)
			return
		}

		s := &stack{name: job.stackName}
		s.setStackName()

		// create session
		sess, err := awsSession()
		if err != nil {
			handleError(err)
			return
		}

		if err := s.change(sess, "rm"); err != nil {
			handleError(err)
		}

	},
}

var list = &cobra.Command{
	Use:   "list",
	Short: "List Change-Sets",
	Run: func(cmd *cobra.Command, args []string) {
		job.request = "change-set list"

		if job.stackName == "" {
			fmt.Println("Please specify stack name using --stack OR -s ...")
			return
		}

		err := configReader(job.cfgFile)
		if err != nil {
			handleError(err)
			return
		}

		s := &stack{name: job.stackName}
		s.setStackName()

		// create session
		sess, err := awsSession()
		if err != nil {
			handleError(err)
			return
		}

		if err := s.change(sess, "list"); err != nil {
			handleError(err)
		}
	},
}

var execute = &cobra.Command{
	Use:   "execute",
	Short: "Execute Change-Set",
	Run: func(cmd *cobra.Command, args []string) {
		job.request = "change-set execute"

		if len(args) < 1 {
			fmt.Println("Please provide Change-Set Name...")
			return
		}

		if job.stackName == "" {
			fmt.Println("Please specify stack name using --stack OR -s ...")
			return
		}

		job.changeName = args[0]

		err := configReader(job.cfgFile)
		if err != nil {
			handleError(err)
			return
		}

		s := &stack{name: job.stackName}
		s.setStackName()

		// create session
		sess, err := awsSession()
		if err != nil {
			handleError(err)
			return
		}

		if err := s.change(sess, "execute"); err != nil {
			handleError(err)
		}
	},
}

var desc = &cobra.Command{
	Use:   "desc",
	Short: "Describe Change-Set",
	Run: func(cmd *cobra.Command, args []string) {
		job.request = "change-set decribe"

		if len(args) < 1 {
			fmt.Println("Please provide Change-Set Name...")
			return
		}

		if job.stackName == "" {
			fmt.Println("Please specify stack name using --stack OR -s ...")
			return
		}

		job.changeName = args[0]

		err := configReader(job.cfgFile)
		if err != nil {
			handleError(err)
			return
		}

		s := &stack{name: job.stackName}
		s.setStackName()

		// create session
		sess, err := awsSession()
		if err != nil {
			handleError(err)
			return
		}

		if err := s.change(sess, "desc"); err != nil {
			handleError(err)
		}
	},
}
