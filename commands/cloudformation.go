package commands

import (
	"fmt"
	"sync"
	"time"

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

// DeployHandler - Handles deploying stacks in the corrcet order
func DeployHandler() {
	// status -  pending, failed, completed
	var status = make(map[string]string)

	for _, stk := range stacks {

		if _, ok := run.stacks[stk.name]; !ok && len(run.stacks) > 0 {
			continue
		}

		// Set deploy status & Check if stack exists
		if stk.stackExists() {
			if err := stk.cleanup(); err != nil {
				Log(fmt.Sprintf("Failed to remove stack: [%s] - %s", stk.name, err.Error()), level.err)
				updateState(status, stk.name, state.failed)
				continue
			}
		}
		updateState(status, stk.name, state.pending)

		if len(stk.dependsOn) == 0 {
			wg.Add(1)
			go func(s stack) {
				defer wg.Done()

				// Deploy 0 Depency Stacks first - each on their on go routine
				Log(fmt.Sprintf("deploying a template for [%s]", s.name), "info")

				if err := s.deploy(); err != nil {
					handleError(err)
				}

				updateState(status, s.name, state.complete)

				// TODO: add deploy logic here
				return
			}(*stk)
			continue
		}

		wg.Add(1)
		go func(s *stack) {
			Log(fmt.Sprintf("[%s] depends on: %s", s.name, s.dependsOn), "info")
			defer wg.Done()

			Log(fmt.Sprintf("Beginning Wait State for Depencies of [%s]"+"\n", s.name), level.debug)
			for {
				depts := []string{}
				for _, dept := range s.dependsOn {
					// Dependency wait
					dp, ok := stacks[dept]
					if !ok {
						Log(fmt.Sprintf("Bad dependency: [%s]", dept), level.err)
						return
					}

					chk, _ := dp.state()

					switch chk {
					case state.failed:
						updateState(status, dp.name, state.failed)
					case state.complete:
						updateState(status, dp.name, state.complete)
					default:
						updateState(status, dp.name, state.pending)
					}

					mutex.Lock()
					depts = append(depts, status[dept])
					mutex.Unlock()
				}

				if all(depts, state.complete) {
					// Deploy stack once dependencies clear
					Log(fmt.Sprintf("Deploying a template for [%s]", s.name), "info")

					if err := s.deploy(); err != nil {
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
		}(stk)

	}

	// Wait for go routines to complete
	wg.Wait()
}

// TerminateHandler - Handles terminating stacks in the correct order
func TerminateHandler() {
	// 	status -  pending, failed, completed
	var status = make(map[string]string)

	for _, stk := range stacks {
		if _, ok := run.stacks[stk.name]; !ok && len(run.stacks) > 0 {
			Log(fmt.Sprintf("%s: not in run.stacks, skipping", stk.name), level.debug)
			continue // only process items in the run.stacks unless empty
		}

		if len(stk.dependsOn) == 0 {
			wg.Add(1)
			go func(s stack) {
				defer wg.Done()
				// Reverse depency look-up so termination waits for all stacks
				// which depend on it, to finish terminating first.
				for {

					for _, stk := range stacks {
						// fmt.Println(stk, stk.dependsOn)
						if stringIn(s.name, stk.dependsOn) {
							Log(fmt.Sprintf("[%s]: Depends on [%s].. Waiting for dependency to terminate", stk.name, s.name), level.info)
							for {

								if !stk.stackExists() {
									break
								}
								time.Sleep(time.Second * 2)
							}
						}
					}

					s.terminate()
					return
				}

			}(*stk)
			continue
		}

		wg.Add(1)
		go func(s *stack) {
			defer wg.Done()

			// Stacks with no Reverse depencies are terminated first
			updateState(status, s.name, state.pending)

			Log(fmt.Sprintf("Terminating stack [%s]", s.stackname), "info")
			if err := s.terminate(); err != nil {
				updateState(status, s.name, state.failed)
				return
			}

			updateState(status, s.name, state.complete)

			return

		}(stk)

	}

	// Wait for go routines to complete
	wg.Wait()
}
