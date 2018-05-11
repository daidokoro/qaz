package stacks

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

var tail chan *TailServiceInput

// TailServiceInput used for tailing cloudfomation outputs
type TailServiceInput struct {
	stk     Stack
	command string
	printed map[string]interface{}
}

// TailService - handles all tailing events
func TailService(tail <-chan *TailServiceInput) {
	Log.Debug("Tail.Service started")

	for input := range tail {
		svc := cloudformation.New(
			input.stk.Session,
			&aws.Config{Credentials: input.stk.creds()},
		)

		params := &cloudformation.DescribeStackEventsInput{
			StackName: aws.String(input.stk.Stackname),
		}

		// If channel is not populated, run verbose cf print
		Log.Debug("calling [DescribeStackEvents] with parameters: %s", params)
		stackevents, err := svc.DescribeStackEvents(params)
		if err != nil {
			Log.Debug("error when tailing events: %v", err)
			continue
		}

		Log.Debug("response: %s", stackevents)

		event := stackevents.StackEvents[0]

		statusReason := ""
		var lg = Log.Info
		if strings.Contains(*event.ResourceStatus, "FAILED") {
			statusReason = *event.ResourceStatusReason
			lg = Log.Error
		}

		line := strings.Join([]string{
			*event.StackName,
			Log.ColorMap(*event.ResourceStatus),
			*event.ResourceType,
			*event.LogicalResourceId,
			statusReason,
		}, " - ")

		if _, ok := input.printed[line]; !ok {
			evt := strings.Split(*event.ResourceStatus, "_")[0]
			if evt == input.command || input.command == "" || strings.Contains(strings.ToLower(evt), "rollback") {
				lg(strings.Trim(line, "- "))
			}

			input.printed[line] = nil
		}
	}

	Log.Debug("Tail.Service closed")
	return
}

// populates tail channel and returns when done
func tailWait(done <-chan bool, tailinput *TailServiceInput) {
	for ch := time.Tick(time.Millisecond * 1300); ; <-ch {
		select {
		case <-done:
			// close(tail)
			return
		default:
			tail <- tailinput
		}
	}
}
