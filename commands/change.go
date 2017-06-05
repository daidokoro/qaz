package commands

import (
	"fmt"
	"qaz/utils"

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

		var s string
		var source string

		if len(args) < 1 {
			fmt.Println("Please provide Change-Set Name...")
			return
		}

		run.changeName = args[0]

		err := configure(run.cfgSource, run.cfgRaw)
		if err != nil {
			utils.HandleError(err)
			return
		}

		if run.tplSource != "" {
			s, source, err = utils.GetSource(run.tplSource)
			if err != nil {
				utils.HandleError(err)
			}
		}

		if len(args) > 0 {
			s = args[0]
		}

		// check if stack exists in config
		if _, ok := stacks[s]; !ok {
			utils.HandleError(fmt.Errorf("Stack [%s] not found in config", s))
		}

		if stacks[s].Source == "" {
			stacks[s].Source = source
		}

		if err = stacks[s].GenTimeParser(); err != nil {
			utils.HandleError(err)
			return
		}

		if err := stacks[s].Change("create", run.changeName); err != nil {
			utils.HandleError(err)
			return
		}

	},
}

var rm = &cobra.Command{
	Use:   "rm",
	Short: "Delete Change-Set",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 1 {
			fmt.Println("Please provide Change-Set Name...")
			return
		}

		if run.stackName == "" {
			fmt.Println("Please specify stack name using --stack OR -s ...")
			return
		}

		run.changeName = args[0]

		err := configure(run.cfgSource, run.cfgRaw)
		if err != nil {
			utils.HandleError(err)
			return
		}

		if _, ok := stacks[run.stackName]; !ok {
			utils.HandleError(fmt.Errorf("Stack not found: [%s]", run.stackName))
		}

		s := stacks[run.stackName]

		if err := s.Change("rm", run.changeName); err != nil {
			utils.HandleError(err)
		}

	},
}

var list = &cobra.Command{
	Use:   "list",
	Short: "List Change-Sets",
	Run: func(cmd *cobra.Command, args []string) {

		if run.stackName == "" {
			fmt.Println("Please specify stack name using --stack OR -s ...")
			return
		}

		err := configure(run.cfgSource, run.cfgRaw)
		if err != nil {
			utils.HandleError(err)
			return
		}

		if _, ok := stacks[run.stackName]; !ok {
			utils.HandleError(fmt.Errorf("Stack not found: [%s]", run.stackName))
		}

		s := stacks[run.stackName]

		if err := s.Change("list", run.changeName); err != nil {
			utils.HandleError(err)
		}
	},
}

var execute = &cobra.Command{
	Use:   "execute",
	Short: "Execute Change-Set",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 1 {
			fmt.Println("Please provide Change-Set Name...")
			return
		}

		if run.stackName == "" {
			fmt.Println("Please specify stack name using --stack OR -s ...")
			return
		}

		run.changeName = args[0]

		err := configure(run.cfgSource, run.cfgRaw)
		if err != nil {
			utils.HandleError(err)
			return
		}

		if _, ok := stacks[run.stackName]; !ok {
			utils.HandleError(fmt.Errorf("Stack not found: [%s]", run.stackName))
		}

		s := stacks[run.stackName]

		if err := s.Change("execute", run.changeName); err != nil {
			utils.HandleError(err)
		}
	},
}

var desc = &cobra.Command{
	Use:   "desc",
	Short: "Describe Change-Set",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 1 {
			fmt.Println("Please provide Change-Set Name...")
			return
		}

		if run.stackName == "" {
			fmt.Println("Please specify stack name using --stack OR -s ...")
			return
		}

		run.changeName = args[0]

		err := configure(run.cfgSource, run.cfgRaw)
		if err != nil {
			utils.HandleError(err)
			return
		}

		if _, ok := stacks[run.stackName]; !ok {
			utils.HandleError(fmt.Errorf("Stack not found: [%s]", run.stackName))
		}

		s := stacks[run.stackName]

		if err := s.Change("desc", run.changeName); err != nil {
			utils.HandleError(err)
		}
	},
}
