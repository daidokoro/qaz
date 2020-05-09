package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/daidokoro/qaz/utils"

	"github.com/spf13/cobra"
)

var lintCmd = &cobra.Command{
	Use:   "lint [stack]",
	Short: "Validates stack by calling cfn-lint",
	Example: strings.Join([]string{
		"",
		"qaz lint -c config.yml -t stack::source",
		"qaz lint vpc -c config.yml",
	}, "\n"),
	PreRun: initialise,
	Run: func(cmd *cobra.Command, args []string) {

		var s string
		var source string

		err := Configure(run.cfgSource, run.cfgRaw)
		utils.HandleError(err)

		switch {
		case run.tplSource != "":
			s, source, err = utils.GetSource(run.tplSource)
			utils.HandleError(err)
		case len(args) > 0:
			s = args[0]
		}

		// check if stack exists in config
		if _, ok := stacks.Get(s); !ok {
			utils.HandleError(fmt.Errorf("Stack [%s] not found in config", s))
			return
		}

		// if source is defined via cli arg
		if source != "" {
			stacks.MustGet(s).Source = source
		}

		name := fmt.Sprintf("%s-%s", project, s)
		log.Debug("generating a template for %s", name)
		utils.HandleError(stacks.MustGet(s).GenTimeParser())

		// write template to temporary file
		content := []byte(stacks.MustGet(s).Template)
		filename := fmt.Sprintf(".%s.qaz", s)
		writeErr := ioutil.WriteFile(filename, content, 0644)
		utils.HandleError(writeErr)

		// run cfn-lint against temporary file
		_, lookErr := exec.LookPath("cfn-lint")
		if lookErr != nil {
			utils.HandleError(fmt.Errorf("cfn-lint executable not found! Please consider https://pypi.org/project/cfn-lint/ for help."))
		}
		execCmd := exec.Command("cfn-lint", filename)
		execCmd.Env = append(os.Environ())
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		execErr := execCmd.Run()
		utils.HandleError(execErr)

	},
}
