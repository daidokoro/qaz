package stacks

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/daidokoro/qaz/log"
)

// Status - Checks stack status, pending, failed, complete
func (s *Stack) Status() error {
	svc := cloudformation.New(s.Session, &aws.Config{Credentials: s.creds()})

	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.Stackname),
	}

	log.Debug("calling [DescribeStacks] with parameters: %s", describeStacksInput)
	status, err := svc.DescribeStacks(describeStacksInput)

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "exist") {
			fmt.Printf("create_pending -> %s [%s]"+"\n", s.Name, s.Stackname)
			return nil
		}
		return err
	}

	// Define time flag
	stat := *status.Stacks[0].StackStatus
	var timeflag time.Time
	switch strings.Split(stat, "_")[0] {
	case "UPDATE":
		timeflag = *status.Stacks[0].LastUpdatedTime
	default:
		timeflag = *status.Stacks[0].CreationTime
	}

	// Print Status
	fmt.Printf(
		"%s%s - %s --> %s - [%s]"+"\n",
		log.ColorString(`@`, "magenta"),
		timeflag.Format(time.RFC850),
		strings.ToLower(log.ColorMap(*status.Stacks[0].StackStatus)),
		s.Name,
		s.Stackname,
	)

	return nil
}
