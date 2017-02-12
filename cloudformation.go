package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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
	built      bool
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

	if _, err := svc.CreateStack(createParams); err != nil {
		log.Fatal("Deploying failed: ", err.Error())
	} else {

		go verbose(s.stackname, "CREATE", session)
		describeStacksInput := &cloudformation.DescribeStacksInput{
			StackName: aws.String(s.stackname),
		}
		if err := svc.WaitUntilStackCreateComplete(describeStacksInput); err != nil {
			// FIXME this works in so far that we wait until the stack is
			// completed and capture errors, but it doesn't really tail
			// cloudroamtion events.
			log.Fatal(err)
		}

		log.Printf("Deployment successful: [%s]"+"\n", s.stackname)
	}

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
		log.Println("Stack exists, updating...")

		_, err := svc.UpdateStack(updateParams)

		if err != nil {
			log.Fatal("Updating stack failed ", err)
			return err
		}

		go verbose(s.stackname, "UPDATE", session)

		describeStacksInput := &cloudformation.DescribeStacksInput{
			StackName: aws.String(s.stackname),
		}

		if err := svc.WaitUntilStackUpdateComplete(describeStacksInput); err != nil {
			// FIXME this works in so far that we wait until the stack is
			// completed and capture errors, but it doesn't really tail
			// cloudroamtion events.
			log.Fatal(err)
		}

		log.Printf("Stack update successful: [%s]"+"\n", s.stackname)

	}
	return nil
}

func (s *stack) terminate(session *session.Session) error {
	svc := cloudformation.New(session)

	params := &cloudformation.DeleteStackInput{
		StackName: aws.String(s.stackname),
	}

	_, err := svc.DeleteStack(params)

	go verbose(s.stackname, "DELETE", session)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			log.Fatal("Deleting failed: ", awsErr.Code(), awsErr.Message())
		} else {
			log.Fatal("Deleting failed ", err)
			return err
		}
	}

	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.stackname),
	}

	if err := svc.WaitUntilStackDeleteComplete(describeStacksInput); err != nil {
		// FIXME this works in so far that we wait until the stack is
		// completed and capture errors, but it doesn't really tail
		// cloudroamtion events.
		log.Fatal(err)
	}

	log.Printf("Deletion successful: [%s]"+"\n", s.stackname)
	return nil
}

func (s *stack) stackExists(session *session.Session) bool {
	svc := cloudformation.New(session)

	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.stackname),
	}

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

	status, err := svc.DescribeStacks(describeStacksInput)

	if err != nil {
		fmt.Printf("create_pending -> %s [%s]"+"\n", s.name, s.stackname)
		s.built = false
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

	s.built = true
	return nil
}

func (s *stack) outputs(session *session.Session) error {

	if s == nil {
		log.Println("Stack does not exist in config")
		return nil
	}

	svc := cloudformation.New(session)

	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.stackname),
	}

	outputs, err := svc.DescribeStacks(describeStacksInput)
	if err != nil {
		log.Println("Unable to reach stack", err.Error())
		return err
	}

	for _, i := range outputs.Stacks {
		fmt.Printf("\n"+"[%s]"+"\n", *i.StackName)
		for _, o := range i.Outputs {
			fmt.Println(*o.OutputKey, "->", *o.OutputValue)
		}
		fmt.Println()
	}

	return nil
}

func (s *stack) state(session *session.Session) (string, error) {
	svc := cloudformation.New(session)

	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.stackname),
	}

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

	resp, err := svc.ValidateTemplate(params)
	if err != nil {
		log.Printf(colorString("Failed!", "red"))
		return err
	}

	log.Printf(
		"%s\n\n%s",
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
				log.Printf("Deploying a template for [%s]", s.name)

				s.deploy(sess)

				updateState(status, s.name, state.complete)

				// TODO: add deploy logic here
				return
			}(*stk, sess)
			continue
		}

		wg.Add(1)
		go func(s *stack, sess *session.Session) {
			log.Printf("[%s] depends on: %s"+"\n", s.name, s.dependsOn)
			defer wg.Done()
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
					log.Printf("Deploying a template for [%s]", s.name)
					s.deploy(sess)
					return
				}

				for _, v := range depts {
					if v == state.failed {
						log.Printf("Error, Deploy Cancelled for stack [%s] due to dependency failure!", s.name)
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

			log.Printf("Terminating stack [%s]", s.stackname)
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
