package stacks

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// tail - tracks the progress during stack updates. c - command Type
func (s *Stack) tail(c string, done <-chan bool) {
	svc := cloudformation.New(s.Session, &aws.Config{Credentials: s.creds()})

	params := &cloudformation.DescribeStackEventsInput{
		StackName: aws.String(s.Stackname),
	}

	// used to track what lines have already been printed, to prevent dubplicate output
	printed := make(map[string]interface{})

	// NOTE: for loop with instant start ticker
	for ch := time.Tick(time.Millisecond * 1300); ; <-ch {
		select {
		case <-done:
			Log.Debug("Tail run.Completed")
			return
		default:
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
