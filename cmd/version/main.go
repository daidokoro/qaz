package cmdVersion

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "v0.82.0-beta"

// NewCmdVersion creates the version command
func NewCmdVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Long:  `Print the version number of the CLI`,
		Run: func(cmd *cobra.Command, args []string) {
			runVersion()
		},
	}

	return cmd
}

func runVersion() {
	fmt.Println(formattedVersion())
}

func formattedVersion() string {
	return fmt.Sprintf("Qaz CLI %s", version)
}



