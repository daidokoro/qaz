package commands

import (
	"fmt"
	"io/ioutil"
	"strings"

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

	log.Debug(fmt.Sprintln("Calling [Invoke] with parameters:", params))
	resp, err := svc.Invoke(params)

	if err != nil {
		return err
	}

	if resp.FunctionError != nil {
		return fmt.Errorf(*resp.FunctionError)
	}

	a.response = string(resp.Payload)

	log.Debug(fmt.Sprintln("Lambda response:", a.response))
	return nil
}

// invoke command
var invokeCmd = &cobra.Command{
	Use:    "invoke",
	Short:  "Invoke AWS Lambda Functions",
	PreRun: initialise,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 1 {
			log.Warn("no lambda function specified")
			return
		}

		sess, err := manager.GetSess(run.profile)
		utils.HandleError(err)

		f := awsLambda{name: args[0]}

		if run.funcEvent != "" {
			if strings.HasPrefix(run.funcEvent, "@") {
				input := strings.Replace(run.funcEvent, "@", "", -1)
				log.Debug(fmt.Sprintf("file input detected [%s], opening file", input))
				event, err := ioutil.ReadFile(input)
				utils.HandleError(err)
				f.payload = event
			} else {
				f.payload = []byte(run.funcEvent)
			}
		}

		if err := f.Invoke(sess); err != nil {
			if strings.Contains(err.Error(), "Unhandled") {
				log.Error(fmt.Sprintf("Unhandled Exception: Potential Issue with Lambda Function Logic: %s\n", f.name))
			}
			utils.HandleError(err)
		}

		fmt.Println(f.response)

	},
}
