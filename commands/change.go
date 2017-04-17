package commands

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
		var s string
		var source string

		if len(args) < 1 {
			fmt.Println("Please provide Change-Set Name...")
			return
		}

		job.changeName = args[0]

		err := configReader(job.cfgSource)
		if err != nil {
			handleError(err)
			return
		}

		if job.tplSource != "" {
			s, source, err = getSource(job.tplSource)
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

		if err = stacks[s].genTimeParser(); err != nil {
			handleError(err)
			return
		}

		if err := stacks[s].change("create"); err != nil {
			handleError(err)
			return
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

		err := configReader(job.cfgSource)
		if err != nil {
			handleError(err)
			return
		}

		if _, ok := stacks[job.stackName]; !ok {
			handleError(fmt.Errorf("Stack not found: [%s]", job.stackName))
		}

		s := stacks[job.stackName]

		if err := s.change("rm"); err != nil {
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

		err := configReader(job.cfgSource)
		if err != nil {
			handleError(err)
			return
		}

		if _, ok := stacks[job.stackName]; !ok {
			handleError(fmt.Errorf("Stack not found: [%s]", job.stackName))
		}

		s := stacks[job.stackName]

		if err := s.change("list"); err != nil {
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

		err := configReader(job.cfgSource)
		if err != nil {
			handleError(err)
			return
		}

		if _, ok := stacks[job.stackName]; !ok {
			handleError(fmt.Errorf("Stack not found: [%s]", job.stackName))
		}

		s := stacks[job.stackName]

		if err := s.change("execute"); err != nil {
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

		err := configReader(job.cfgSource)
		if err != nil {
			handleError(err)
			return
		}

		if _, ok := stacks[job.stackName]; !ok {
			handleError(fmt.Errorf("Stack not found: [%s]", job.stackName))
		}

		s := stacks[job.stackName]

		if err := s.change("desc"); err != nil {
			handleError(err)
		}
	},
}
