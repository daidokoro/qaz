go
package cmdRoot

import (
	"qaz/cmd/aws_lambda"
	"qaz/cmd/change"
	"qaz/cmd/completion"
	"qaz/cmd/config"
	"qaz/cmd/deploy"
	"qaz/cmd/generate"
	"qaz/cmd/init"
	"qaz/cmd/lint"
	"qaz/cmd/outputs"
	"qaz/cmd/sessions"
	"qaz/cmd/shell"
	"qaz/cmd/version"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "qaz",
	Short: "Qaz is a CLI tool for managing cloud resources.",
	Long: `Qaz is a CLI tool designed to simplify the management
of cloud resources across various providers.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func addCommands() {
	rootCmd.AddCommand(cmdDeploy.NewCmdDeploy())
	rootCmd.AddCommand(cmdConfig.NewCmdConfig())
	rootCmd.AddCommand(cmdInit.NewCmdInit())
	rootCmd.AddCommand(cmdVersion.NewCmdVersion())
	rootCmd.AddCommand(cmdLint.NewCmdLint())
	rootCmd.AddCommand(cmdGenerate.NewCmdGenerate())
	rootCmd.AddCommand(cmdChange.NewCmdChange())
	rootCmd.AddCommand(cmdOutputs.NewCmdOutputs())
	rootCmd.AddCommand(cmdSessions.NewCmdSessions())
	rootCmd.AddCommand(cmdCompletion.NewCmdCompletion())
	rootCmd.AddCommand(cmdAwsLambda.NewCmdAwsLambda())
	rootCmd.AddCommand(cmdShell.NewCmdShell())
}

func init() {
	addCommands()
}

