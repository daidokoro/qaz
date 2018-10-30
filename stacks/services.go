package stacks

import (
	"context"
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
	log.Debug("Tail.Service started")

	for input := range tail {
		svc := cloudformation.New(
			input.stk.Session,
			&aws.Config{Credentials: input.stk.creds()},
		)

		params := &cloudformation.DescribeStackEventsInput{
			StackName: aws.String(input.stk.Stackname),
		}

		// If channel is not populated, run verbose cf print
		log.Debug("calling [DescribeStackEvents] with parameters: %s", params)
		stackevents, err := svc.DescribeStackEvents(params)
		if err != nil {
			log.Debug("error when tailing events: %v", err)
			continue
		}

		log.Debug("response: %s", stackevents)

		event := stackevents.StackEvents[0]

		statusReason := ""
		var lg = log.Info
		if strings.Contains(*event.ResourceStatus, "FAILED") {
			statusReason = *event.ResourceStatusReason
			lg = log.Error
		}

		line := strings.Join([]string{
			*event.StackName,
			log.ColorMap(*event.ResourceStatus),
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

	log.Debug("Tail.Service closed")
	return
}

// populates tail channel and returns when done
func tailWait(ctx context.Context, tailinput *TailServiceInput) {
	for ch := time.Tick(time.Millisecond * 1300); ; <-ch {
		select {
		case <-ctx.Done():
			// close(tail)
			return
		default:
			tail <- tailinput
		}
	}
}
