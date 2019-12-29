package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/daidokoro/qaz/utils"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [shell]",
	Short: "Output shell completion code for the specified shell (bash or zsh)",
	Example: strings.Join([]string{
		"",
		"qaz completion bash > $(brew --prefix)/etc/bash_completion.d/qaz",
		"qaz completion zsh > \"${fpath[1]}/_qaz\"",
	}, "\n"),
	PreRun: initialise,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			utils.HandleError(fmt.Errorf("please specify shell"))

		} else if len(args) > 1 {
			utils.HandleError(fmt.Errorf("too many arguments, expected shell type only"))

		} else {

			if args[0] == "bash" {
				RootCmd.GenBashCompletion(os.Stdout);
			} else if args[0] == "zsh" {
				RootCmd.GenZshCompletion(os.Stdout);
			} else {
				utils.HandleError(fmt.Errorf("Shell [%s] currently not supported", args[0]))
			}

		}

	},
}

