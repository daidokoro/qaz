package commands

import (
	"encoding/json"
	"fmt"
	"github.com/daidokoro/qaz/utils"

	stks "github.com/daidokoro/qaz/stacks"

	"github.com/spf13/cobra"
)

// output and export commands

var (
	// output command
	outputsCmd = &cobra.Command{
		Use:     "outputs [stack]",
		Short:   "Prints stack outputs",
		Example: "qaz outputs vpc subnets --config path/to/config",
		PreRun:  initialise,
		Run: func(cmd *cobra.Command, args []string) {

			if len(args) < 1 {
				fmt.Println("Please specify stack(s) to check, For details try --> qaz outputs --help")
				return
			}

			err := Configure(run.cfgSource, run.cfgRaw)
			utils.HandleError(err)

			for _, s := range args {
				// check if stack exists
				if _, ok := stacks[s]; !ok {
					utils.HandleError(fmt.Errorf("%s: does not Exist in Config", s))
				}

				wg.Add(1)
				go func(s string) {
					if err := stacks[s].Outputs(); err != nil {
						log.Error(err.Error())
						wg.Done()
						return
					}

					for _, i := range stacks[s].Output.Stacks {
						m, err := json.MarshalIndent(i.Outputs, "", "  ")
						if err != nil {
							log.Error(err.Error())
						}
						fmt.Println(string(m))
					}

					wg.Done()
				}(s)
			}
			wg.Wait()

		},
	}

	// export command
	exportsCmd = &cobra.Command{
		Use:     "exports",
		Short:   "Prints stack exports",
		Example: "qaz exports",
		PreRun:  initialise,
		Run: func(cmd *cobra.Command, args []string) {

			sess, err := manager.GetSess(run.profile)
			utils.HandleError(err)

			err = stks.Exports(sess)
			utils.HandleError(err)

		},
	}
)
