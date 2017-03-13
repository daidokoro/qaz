package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// stack - holds all meaningful information about a particular stack.
type stack struct {
	name         string
	stackname    string
	template     string
	dependsOn    []interface{}
	dependents   []interface{}
	stackoutputs *cloudformation.DescribeStacksOutput
	parameters   []*cloudformation.Parameter
}

// setStackName - sets the stackname with struct
func (s *stack) setStackName() {
	s.stackname = fmt.Sprintf("%s-%s", project, s.name)
}

func (s *stack) deploy(session *session.Session) error {

	err := s.deployTimeParser()
	if err != nil {
		return err
	}

	Log(fmt.Sprintf("Updated Template:\n%s", s.template), level.debug)

	svc := cloudformation.New(session)

	createParams := &cloudformation.CreateStackInput{
		StackName:       aws.String(s.stackname),
		DisableRollback: aws.Bool(true), // no rollback by default
		TemplateBody:    aws.String(s.template),
	}

	// NOTE: Add parameters flag here if params set
	if len(s.parameters) > 0 {
		createParams.Parameters = s.parameters
	}

	// If IAM is bening touched, add Capabilities
	if strings.Contains(s.template, "AWS::IAM") {
		createParams.Capabilities = []*string{
			aws.String(cloudformation.CapabilityCapabilityIam),
		}
	}

	Log(fmt.Sprintln("Calling [CreateStack] with parameters:", createParams), level.debug)
	if _, err := svc.CreateStack(createParams); err != nil {
		return errors.New(fmt.Sprintln("Deploying failed: ", err.Error()))

	}

	go verbose(s.stackname, "CREATE", session)
	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.stackname),
	}

	Log(fmt.Sprintln("Calling [WaitUntilStackCreateComplete] with parameters:", describeStacksInput), level.debug)
	if err := svc.WaitUntilStackCreateComplete(describeStacksInput); err != nil {
		return err
	}

	Log(fmt.Sprintf("Deployment successful: [%s]", s.stackname), "info")

	return nil
}

func (s *stack) update(session *session.Session) error {
	svc := cloudformation.New(session)
	capability := "CAPABILITY_IAM"
	updateParams := &cloudformation.UpdateStackInput{
		StackName:    aws.String(s.stackname),
		TemplateBody: aws.String(s.template),
		Capabilities: []*string{&capability},
	}

	// NOTE: Add parameters flag here if params set
	if len(s.parameters) > 0 {
		updateParams.Parameters = s.parameters
	}

	if s.stackExists(session) {
		Log("Stack exists, updating...", "info")

		Log(fmt.Sprintln("Calling [UpdateStack] with parameters:", updateParams), level.debug)
		_, err := svc.UpdateStack(updateParams)

		if err != nil {
			return errors.New(fmt.Sprintln("Update failed: ", err))
		}

		go verbose(s.stackname, "UPDATE", session)

		describeStacksInput := &cloudformation.DescribeStacksInput{
			StackName: aws.String(s.stackname),
		}
		Log(fmt.Sprintln("Calling [WaitUntilStackUpdateComplete] with parameters:", describeStacksInput), level.debug)
		if err := svc.WaitUntilStackUpdateComplete(describeStacksInput); err != nil {
			return err
		}

		Log(fmt.Sprintf("Stack update successful: [%s]", s.stackname), "info")

	}
	return nil
}

func (s *stack) terminate(session *session.Session) error {
	svc := cloudformation.New(session)

	params := &cloudformation.DeleteStackInput{
		StackName: aws.String(s.stackname),
	}

	Log(fmt.Sprintln("Calling [DeleteStack] with parameters:", params), level.debug)
	_, err := svc.DeleteStack(params)

	go verbose(s.stackname, "DELETE", session)

	if err != nil {
		return errors.New(fmt.Sprintln("Deleting failed: ", err))
	}

	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.stackname),
	}

	Log(fmt.Sprintln("Calling [WaitUntilStackDeleteComplete] with parameters:", describeStacksInput), level.debug)
	if err := svc.WaitUntilStackDeleteComplete(describeStacksInput); err != nil {
		return err
	}

	Log(fmt.Sprintf("Deletion successful: [%s]", s.stackname), "info")
	return nil
}

func (s *stack) stackExists(session *session.Session) bool {
	svc := cloudformation.New(session)

	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.stackname),
	}

	Log(fmt.Sprintln("Calling [DescribeStacks] with parameters:", describeStacksInput), level.debug)
	_, err := svc.DescribeStacks(describeStacksInput)

	if err == nil {
		return true
	}

	return false
}

func (s *stack) status(session *session.Session) error {
	svc := cloudformation.New(session)

	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.stackname),
	}

	Log(fmt.Sprintln("Calling [UpdateStack] with parameters:", describeStacksInput), level.debug)
	status, err := svc.DescribeStacks(describeStacksInput)

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "exist") {
			fmt.Printf("create_pending -> %s [%s]"+"\n", s.name, s.stackname)
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
		colorString(`@`, "magenta"),
		timeflag.Format(time.RFC850),
		strings.ToLower(colorMap(*status.Stacks[0].StackStatus)),
		s.name,
		s.stackname,
	)

	return nil
}

func (s *stack) state(session *session.Session) (string, error) {
	svc := cloudformation.New(session)

	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.stackname),
	}

	Log(fmt.Sprintln("Calling [DescribeStacks] with parameters: ", describeStacksInput), level.debug)
	status, err := svc.DescribeStacks(describeStacksInput)
	if err != nil {
		if strings.Contains(err.Error(), "not exist") {
			return state.pending, nil
		}
		return "", err
	}

	if strings.Contains(strings.ToLower(status.GoString()), "complete") {
		return state.complete, nil
	} else if strings.Contains(strings.ToLower(status.GoString()), "fail") {
		return state.failed, nil
	}
	return "", nil
}

func (s *stack) change(session *session.Session, req string) error {
	svc := cloudformation.New(session)

	switch req {

	case "create":
		// Resolve Deploy-Time functions
		err := s.deployTimeParser()
		if err != nil {
			return err
		}

		params := &cloudformation.CreateChangeSetInput{
			StackName:     aws.String(s.stackname),
			ChangeSetName: aws.String(job.changeName),
		}

		Log(fmt.Sprintf("Updated Template:\n%s", s.template), level.debug)

		params.TemplateBody = aws.String(s.template)
		if _, err = svc.CreateChangeSet(params); err != nil {
			return err
		}

		describeParams := &cloudformation.DescribeChangeSetInput{
			StackName:     aws.String(s.stackname),
			ChangeSetName: aws.String(job.changeName),
		}

		for {
			// Waiting for PENDING state to change
			resp, err := svc.DescribeChangeSet(describeParams)
			if err != nil {
				return err
			}

			Log(fmt.Sprintf("Creating Change-Set: [%s] - %s - %s", job.changeName, colorMap(*resp.Status), s.stackname), level.info)

			if *resp.Status == "CREATE_COMPLETE" || *resp.Status == "FAILED" {
				break
			}

			time.Sleep(time.Second * 1)
		}

	case "rm":
		params := &cloudformation.DeleteChangeSetInput{
			ChangeSetName: aws.String(job.changeName),
			StackName:     aws.String(s.stackname),
		}

		if _, err := svc.DeleteChangeSet(params); err != nil {
			return err
		}

		Log(fmt.Sprintf("Change-Set: [%s] deleted", job.changeName), level.info)

	case "list":
		params := &cloudformation.ListChangeSetsInput{
			StackName: aws.String(s.stackname),
		}

		resp, err := svc.ListChangeSets(params)
		if err != nil {
			return err
		}

		// if strings.Contains(resp.GoString(), "Summaries:") {
		for _, i := range resp.Summaries {
			Log(fmt.Sprintf("%s%s - Change-Set: [%s] - Status: [%s]", colorString("@", "magenta"), i.CreationTime.Format(time.RFC850), *i.ChangeSetName, *i.ExecutionStatus), level.info)
		}
		// }

	case "execute":
		params := &cloudformation.ExecuteChangeSetInput{
			StackName:     aws.String(s.stackname),
			ChangeSetName: aws.String(job.changeName),
		}

		if _, err := svc.ExecuteChangeSet(params); err != nil {
			return err
		}

		describeStacksInput := &cloudformation.DescribeStacksInput{
			StackName: aws.String(s.stackname),
		}

		go verbose(s.stackname, "UPDATE", session)

		Log(fmt.Sprintln("Calling [WaitUntilStackUpdateComplete] with parameters:", describeStacksInput), level.debug)
		if err := svc.WaitUntilStackUpdateComplete(describeStacksInput); err != nil {
			return err
		}

	case "desc":
		params := &cloudformation.DescribeChangeSetInput{
			ChangeSetName: aws.String(job.changeName),
			StackName:     aws.String(s.stackname),
		}

		resp, err := svc.DescribeChangeSet(params)
		if err != nil {
			return err
		}

		o, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", o)
	}

	return nil
}
