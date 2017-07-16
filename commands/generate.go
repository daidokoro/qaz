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
		if err != nil {
			utils.HandleError(err)
			return
		}

		if run.tplSource != "" {
			s, source, err = utils.GetSource(run.tplSource)
			if err != nil {
				utils.HandleError(err)
				return
			}
		}

		if len(args) > 0 {
			s = args[0]
		}

		// check if stack exists in config
		if _, ok := stacks[s]; !ok {
			utils.HandleError(fmt.Errorf("Stack [%s] not found in config", s))
			return
		}

		if stacks[s].Source == "" {
			stacks[s].Source = source
		}

		name := fmt.Sprintf("%s-%s", project, s)
		log.Debug(fmt.Sprintln("Generating a template for ", name))

		err = stacks[s].GenTimeParser()
		if err != nil {
			utils.HandleError(err)
			return
		}

		reg, err := regexp.Compile(OutputRegex)
		utils.HandleError(err)

		resp := reg.ReplaceAllStringFunc(string(stacks[s].Template), func(s string) string {
			return log.ColorString(s, "cyan")
		})

		fmt.Println(resp)
	},
}
