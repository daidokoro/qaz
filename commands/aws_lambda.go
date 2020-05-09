package commands

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/daidokoro/qaz/log"
	"github.com/daidokoro/qaz/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/spf13/cobra"
)

type awsLambda struct {
	name     string
	payload  []byte
	response string
}

func (a *awsLambda) Invoke(sess *session.Session) error {
	svc := lambda.New(sess)

	params := &lambda.InvokeInput{
		FunctionName: aws.String(a.name),
	}

	if a.payload != nil {
		params.Payload = a.payload
	}

	if run.lambdAsync {
		params.InvocationType = aws.String("Event")
	}

	log.Debug("Calling [Invoke] with parameters: %s", params)
	resp, err := svc.Invoke(params)
	if err != nil {
		return err
	}

	if resp.FunctionError != nil {
		return fmt.Errorf(*resp.FunctionError)
	}

	a.response = string(resp.Payload)

	if run.lambdAsync {
		code := strconv.FormatInt(*resp.StatusCode, 10)
		log.Info("lambda async response code: %s", log.ColorString(code, "green"))
		return nil
	}

	log.Debug("lambda response: %s", a.response)
	return nil
}

// invoke command
var invokeCmd = &cobra.Command{
	Use:     "invoke",
	Short:   "Invoke AWS Lambda Functions [To be DEPRACATED in a future release]",
	Example: "qaz invoke some_function --event @path/to/event.json",
	PreRun:  initialise,
	Run: func(cmd *cobra.Command, args []string) {
		log.Warn("this feature will be DEPRACATED in the next release...")

		if len(args) < 1 {
			log.Warn("no lambda function specified")
			return
		}

		sess, err := GetSession()
		utils.HandleError(err)

		f := awsLambda{name: args[0]}

		if run.funcEvent != "" {
			if strings.HasPrefix(run.funcEvent, "@") {
				input := strings.Replace(run.funcEvent, "@", "", -1)
				log.Debug("file input detected [%s], opening file", input)
				event, err := ioutil.ReadFile(input)
				utils.HandleError(err)
				f.payload = event
			} else {
				f.payload = []byte(run.funcEvent)
			}
		}

		if err := f.Invoke(sess); err != nil {
			if strings.Contains(err.Error(), "Unhandled") {
				log.Error("Unhandled Exception: Potential Issue with Lambda Function Logic: %s\n", f.name)
			}
			utils.HandleError(err)
		}

		if !run.lambdAsync {
			fmt.Println(f.response)
		}
		return
	},
}
