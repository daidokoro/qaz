package stacks

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// tail - tracks the progress during stack updates. c - command Type
func (s *Stack) tail(ctx context.Context, c string) {
	svc := cloudformation.New(s.Session, &aws.Config{Credentials: s.creds()})

	params := &cloudformation.DescribeStackEventsInput{
		StackName: aws.String(s.Stackname),
	}

	// used to track what lines have already been printed, to prevent dubplicate output
	printed := make(map[string]interface{})

	// NOTE: for loop with instant start ticker
	for ch := time.Tick(time.Millisecond * 1300); ; <-ch {
		select {
		case <-ctx.Done():
			log.Debug("Tail run.Completed")
			return
		default:
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

			if _, ok := printed[line]; !ok {
				evt := strings.Split(*event.ResourceStatus, "_")[0]
				if evt == c || c == "" || strings.Contains(strings.ToLower(evt), "rollback") {
					lg(strings.Trim(line, "- "))
				}

				printed[line] = nil
			}

		}

	}
}
