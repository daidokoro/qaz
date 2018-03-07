package commands

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/daidokoro/qaz/utils"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate [stack]",
	Short: "Generates template from configuration values",
	Example: strings.Join([]string{
		"",
		"qaz generate -c config.yml -t stack::source",
		"qaz generate vpc -c config.yml",
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

		resp := regexp.MustCompile(OutputRegex).
			ReplaceAllStringFunc(string(stacks.MustGet(s).Template), func(s string) string {
				return log.ColorString(s, "cyan")
			})

		fmt.Println(resp)
	},
}
