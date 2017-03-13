package main

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

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

// StackOutputs - Returns outputs of given stackname
func StackOutputs(name string, session *session.Session) (*cloudformation.DescribeStacksOutput, error) {

	svc := cloudformation.New(session)
	outputParams := &cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	}

	Log(fmt.Sprintln("Calling [DescribeStacks] with parameters:", outputParams), level.debug)
	outputs, err := svc.DescribeStacks(outputParams)
	if err != nil {
		return &cloudformation.DescribeStacksOutput{}, errors.New(fmt.Sprintln("Unable to reach stack", err.Error()))
	}

	return outputs, nil
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

		if _, ok := job.stacks[stk.name]; !ok && len(job.stacks) > 0 {
			continue
		}

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

				if err := s.deploy(sess); err != nil {
					handleError(err)
				}

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

					if err := s.deploy(sess); err != nil {
						handleError(err)
					}
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
	// 	status -  pending, failed, completed
	var status = make(map[string]string)

	sess, _ := awsSession()

	for _, stk := range stacks {
		if _, ok := job.stacks[stk.name]; !ok && len(job.stacks) > 0 {
			Log(fmt.Sprintf("%s: not in job.stacks, skipping", stk.name), level.debug)
			continue // only process items in the job.stacks unless empty
		}

		if len(stk.dependsOn) == 0 {
			wg.Add(1)
			go func(s stack, sess *session.Session) {
				defer wg.Done()
				// Reverse depency look-up so termination waits for all stacks
				// which depend on it, to finish terminating first.
				for {

					for _, stk := range stacks {
						// fmt.Println(stk, stk.dependsOn)
						if stringIn(s.name, stk.dependsOn) {
							Log(fmt.Sprintf("[%s]: Depends on [%s].. Waiting for dependency to terminate", stk.name, s.name), level.info)
							for {

								if !stk.stackExists(sess) {
									break
								}
								time.Sleep(time.Second * 2)
							}
						}
					}

					s.terminate(sess)
					return
				}

			}(*stk, sess)
			continue
		}

		wg.Add(1)
		go func(s *stack, sess *session.Session) {
			defer wg.Done()

			// Stacks with no Reverse depencies are terminated first
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
