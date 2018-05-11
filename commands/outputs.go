package commands

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sync"

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
			var wg sync.WaitGroup
			if len(args) < 1 {
				fmt.Println("Please specify stack(s) to check, For details try --> qaz outputs --help")
				return
			}

			err := Configure(run.cfgSource, run.cfgRaw)
			utils.HandleError(err)

			for _, s := range args {
				// check if stack exists
				if _, ok := stacks.Get(s); !ok {
					utils.HandleError(fmt.Errorf("%s: does not Exist in Config", s))
				}

				wg.Add(1)
				go func(s string) {
					defer wg.Done()
					if err := stacks.MustGet(s).Outputs(); err != nil {
						log.Error(err.Error())
						return
					}

					for _, i := range stacks.MustGet(s).Output.Stacks {
						m, err := json.MarshalIndent(i.Outputs, "", "  ")
						if err != nil {
							log.Error(err.Error())
						}

						resp := regexp.MustCompile(OutputRegex).
							ReplaceAllStringFunc(string(m), func(s string) string {
								return log.ColorString(s, "cyan")
							})

						fmt.Println(resp)
					}
					return
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
			sess, err := GetSession()
			utils.HandleError(err)
			utils.HandleError(stks.Exports(sess))
		},
	}
)
