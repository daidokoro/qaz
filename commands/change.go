package commands

import (
	"fmt"

	"github.com/daidokoro/qaz/log"
	"github.com/daidokoro/qaz/utils"

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
	Use:    "create",
	Short:  "Create Changet-Set",
	PreRun: initialise,
	Run: func(cmd *cobra.Command, args []string) {

		var s string
		var source string

		if len(args) < 1 {
			fmt.Println("Please provide Change-Set Name...")
			return
		}

		if run.stackName == "" && run.tplSource == "" {
			fmt.Println("Please specify stack name using --stack, -s  or -t, --template...")
			return
		}

		run.changeName = args[0]

		err := Configure(run.cfgSource, run.cfgRaw)
		utils.HandleError(err)

		if run.tplSource != "" {
			s, source, err = utils.GetSource(run.tplSource)
			utils.HandleError(err)
		}

		if run.stackName != "" && s == "" {
			s = run.stackName
		}

		// check if stack exists in config
		if _, ok := stacks.Get(s); !ok {
			utils.HandleError(fmt.Errorf("Stack [%s] not found in config", s))
		}

		if stacks.MustGet(s).Source == "" {
			stacks.MustGet(s).Source = source
		}

		err = stacks.MustGet(s).GenTimeParser()
		utils.HandleError(err)

		err = stacks.MustGet(s).Change("create", run.changeName)
		utils.HandleError(err)

		log.Info("change-set [%s] creation successful", run.changeName)

	},
}

var rm = &cobra.Command{
	Use:    "rm",
	Short:  "Delete Change-Set",
	PreRun: initialise,
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

		err := Configure(run.cfgSource, run.cfgRaw)
		utils.HandleError(err)

		if _, ok := stacks.Get(run.stackName); !ok {
			utils.HandleError(fmt.Errorf("Stack not found: [%s]", run.stackName))
		}

		err = stacks.MustGet(run.stackName).Change("rm", run.changeName)
		utils.HandleError(err)

	},
}

var list = &cobra.Command{
	Use:    "list",
	Short:  "List Change-Sets",
	PreRun: initialise,
	Run: func(cmd *cobra.Command, args []string) {

		if run.stackName == "" {
			fmt.Println("Please specify stack name using --stack OR -s ...")
			return
		}

		err := Configure(run.cfgSource, run.cfgRaw)
		utils.HandleError(err)

		if _, ok := stacks.Get(run.stackName); !ok {
			utils.HandleError(fmt.Errorf("Stack not found: [%s]", run.stackName))
		}

		err = stacks.MustGet(run.stackName).Change("list", run.changeName)
		utils.HandleError(err)
	},
}

var execute = &cobra.Command{
	Use:    "execute",
	Short:  "Execute Change-Set",
	PreRun: initialise,
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

		err := Configure(run.cfgSource, run.cfgRaw)
		if err != nil {
			utils.HandleError(err)
			return
		}

		if _, ok := stacks.Get(run.stackName); !ok {
			utils.HandleError(fmt.Errorf("Stack not found: [%s]", run.stackName))
		}

		err = stacks.MustGet(run.stackName).Change("execute", run.changeName)
		utils.HandleError(err)

		log.Info("change-set [%s] execution successful", run.changeName)
	},
}

var desc = &cobra.Command{
	Use:    "desc",
	Short:  "Describe Change-Set",
	PreRun: initialise,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 1 {
			fmt.Println("please provide Change-Set name")
			return
		}

		if run.stackName == "" {
			fmt.Println("Please specify stack name using --stack OR -s ...")
			return
		}

		run.changeName = args[0]

		err := Configure(run.cfgSource, run.cfgRaw)
		if err != nil {
			utils.HandleError(err)
			return
		}

		if _, ok := stacks.Get(run.stackName); !ok {
			utils.HandleError(fmt.Errorf("Stack not found: [%s]", run.stackName))
		}

		err = stacks.MustGet(run.stackName).Change("desc", run.changeName)
		utils.HandleError(err)
	},
}
