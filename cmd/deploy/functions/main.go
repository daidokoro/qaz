package cmdFunctions

import (
	"fmt"

	"github.com/daidokoro/qaz/utils"
	"github.com/spf13/cobra"
)

// functionsCmd represents the functions command
var functionsCmd = &cobra.Command{
	Use:   "functions",
	Short: "Manage Functions Deployed with your stack",
	Long: `
functions command enables you to directly interact with functions deployed 
with your stack.

Currently functions only support Invoking`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("functions called")
	},
}

func NewCmdFunctions() *cobra.Command {
	return functionsCmd
}

func init() {
	err := utils.CheckAWS()
	utils.HandleError(err)
}