package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// stack - holds all meaningful information about a particular stack.
type stack struct {
	name       string
	stackname  string
	template   string
	dependsOn  []interface{}
	dependents []interface{}
}

// State - struct for handling stack deploy/terminate statuses
var state = struct {
	pending  string
	failed   string
	complete string
}{
	complete: "complete",
	pending:  "pending",
	failed:   "failed",
}

// mutex - used to sync access to cross thread variables
var mutex = &sync.Mutex{}

// updateState - Locks cross channel object and updates value
func updateState(statusMap map[string]string, name string, status string) {
	Log(fmt.Sprintf("Updating Stack Status Map: %s - %s", name, status), level.debug)
	mutex.Lock()
	statusMap[name] = status
	mutex.Unlock()
}

// setStackName - sets the stackname with struct
func (s *stack) setStackName() {
	s.stackname = fmt.Sprintf("%s-%s", project, s.name)
}

func (s *stack) deploy(session *session.Session) error {
	svc := cloudformation.New(session)
	capability := "CAPABILITY_IAM"
	createParams := &cloudformation.CreateStackInput{
		StackName:       aws.String(s.stackname),
		DisableRollback: aws.Bool(true), // no rollback by default
		TemplateBody:    aws.String(s.template),
		Capabilities:    []*string{&capability},
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
		// FIXME this works in so far that we wait until the stack is
		// completed and capture errors, but it doesn't really tail
		// cloudroamtion events.
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
			// FIXME this works in so far that we wait until the stack is
			// completed and capture errors, but it doesn't really tail
			// cloudroamtion events.
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
		// FIXME this works in so far that we wait until the stack is
		// completed and capture errors, but it doesn't really tail
		// cloudroamtion events.
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

func (s *stack) outputs(session *session.Session) error {

	if s == nil {
		Log("Stack does not exist in config", "warn")
		return nil
	}

	svc := cloudformation.New(session)

	outputParams := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.stackname),
	}

	Log(fmt.Sprintln("Calling [DescribeStacks] with parameters:", outputParams), level.debug)
	outputs, err := svc.DescribeStacks(outputParams)
	if err != nil {
		return errors.New(fmt.Sprintln("Unable to reach stack", err.Error()))
	}

	for _, i := range outputs.Stacks {

		fmt.Printf("\n"+"[%s]"+"\n", *i.StackName)
		for _, o := range i.Outputs {

			fmt.Printf("  Description: %s\n  %s: %s\n\n", *o.Description, colorString(*o.OutputKey, "magenta"), *o.OutputValue)

		}
		fmt.Println()

	}

	return nil
}

// Exports - prints all cloudformation exports
func Exports(session *session.Session) error {

	svc := cloudformation.New(session)

	exportParams := &cloudformation.ListExportsInput{}

	Log(fmt.Sprintln("Calling [ListExports] with parameters:", exportParams), level.debug)
	exports, err := svc.ListExports(exportParams)

	if err != nil {
		return err
	}

	for _, i := range exports.Exports {

		fmt.Printf("Export Name: %s\nExport Value: %s\n--\n", colorString(*i.Name, "magenta"), *i.Value)
	}

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

// Check - Validates Cloudformation Templates
func Check(template string, session *session.Session) error {
	svc := cloudformation.New(session)
	params := &cloudformation.ValidateTemplateInput{
		TemplateBody: aws.String(template),
	}

	Log(fmt.Sprintf("Calling [ValidateTemplate] with parameters:\n%s"+"\n--\n", params), level.debug)
	resp, err := svc.ValidateTemplate(params)
	if err != nil {
		return err
	}

	fmt.Printf(
		"%s\n\n%s"+"\n",
		colorString("Valid!", "green"),
		resp.GoString(),
	)

	return nil

}

// DeployHandler - Handles deploying stacks in the corrcet order
func DeployHandler() {
	// status -  pending, failed, completed
	var status = make(map[string]string)

	sess, _ := awsSession()

	for _, stk := range stacks {
		// Set deploy status & Check if stack exists
		if stk.stackExists(sess) {

			updateState(status, stk.name, state.complete)
			fmt.Printf("Stack [%s] already exists..."+"\n", stk.name)
			continue
		} else {
			updateState(status, stk.name, state.pending)
		}

		if len(stk.dependsOn) == 0 {
			wg.Add(1)
			go func(s stack, sess *session.Session) {
				defer wg.Done()

				// Deploy 0 Depency Stacks first - each on their on go routine
				Log(fmt.Sprintf("Deploying a template for [%s]", s.name), "info")

				s.deploy(sess)

				updateState(status, s.name, state.complete)

				// TODO: add deploy logic here
				return
			}(*stk, sess)
			continue
		}

		wg.Add(1)
		go func(s *stack, sess *session.Session) {
			Log(fmt.Sprintf("[%s] depends on: %s", s.name, s.dependsOn), "info")
			defer wg.Done()

			Log(fmt.Sprintf("Beginning Wait State for Depencies of [%s]"+"\n", s.name), level.debug)
			for {
				depts := []string{}
				for _, dept := range s.dependsOn {
					// Dependency wait
					dp := &stack{name: dept.(string)}
					dp.setStackName()
					chk, _ := dp.state(sess)

					switch chk {
					case state.failed:
						updateState(status, dp.name, state.failed)
					case state.complete:
						updateState(status, dp.name, state.complete)
					default:
						updateState(status, dp.name, state.pending)
					}

					mutex.Lock()
					depts = append(depts, status[dept.(string)])
					mutex.Unlock()
				}

				if all(depts, state.complete) {
					// Deploy stack once dependencies clear
					Log(fmt.Sprintf("Deploying a template for [%s]", s.name), "info")
					s.deploy(sess)
					return
				}

				for _, v := range depts {
					if v == state.failed {
						Log(fmt.Sprintf("Deploy Cancelled for stack [%s] due to dependency failure!", s.name), "warn")
						return
					}
				}

				time.Sleep(time.Second * 1)
			}
		}(stk, sess)

	}

	// Wait for go routines to complete
	wg.Wait()
}

// TerminateHandler - Handles terminating stacks in the correct order
func TerminateHandler() {
	// status -  pending, failed, completed
	var status = make(map[string]string)

	sess, _ := awsSession()

	for _, stk := range stacks {
		// Check if stack exists

		if len(stk.dependsOn) == 0 {
			wg.Add(1)
			go func(s stack, sess *session.Session) {
				defer wg.Done()
				// Reverse depency look-up so termination waits for all stacks
				// which depend on it, to finish terminating first.
				for {
					depts := []string{}

					for _, stk := range stacks {

						if stringIn(s.name, stk.dependsOn) {

							mutex.Lock()
							depts = append(depts, status[stk.name])
							mutex.Unlock()
						}
					}

					if all(depts, state.complete) {
						s.terminate(sess)
						updateState(status, s.name, state.complete)
						return
					}

				}

			}(*stk, sess)
			continue
		}

		wg.Add(1)
		go func(s *stack, sess *session.Session) {
			defer wg.Done()

			// Stacks with no Reverse depencies are terminate first
			updateState(status, s.name, state.pending)

			Log(fmt.Sprintf("Terminating stack [%s]", s.stackname), "info")
			if err := s.terminate(sess); err != nil {
				updateState(status, s.name, state.failed)
				return
			}

			updateState(status, s.name, state.complete)

			return

		}(stk, sess)

	}

	// Wait for go routines to complete
	wg.Wait()
}
