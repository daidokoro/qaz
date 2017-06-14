package stacks

import (
	"fmt"
	"qaz/utils"
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
	Log.Debug(fmt.Sprintf("Updating Stack Status Map: %s - %s", name, status))
	mutex.Lock()
	statusMap[name] = status
	mutex.Unlock()
}

// Exports - prints all cloudformation exports
func Exports(session *session.Session) error {

	svc := cloudformation.New(session)

	exportParams := &cloudformation.ListExportsInput{}

	Log.Debug(fmt.Sprintln("Calling [ListExports] with parameters:", exportParams))
	exports, err := svc.ListExports(exportParams)

	if err != nil {
		return err
	}

	for _, i := range exports.Exports {

		fmt.Printf("Export Name: %s\nExport Value: %s\n--\n", Log.ColorString(*i.Name, "magenta"), *i.Value)
	}

	return nil
}

// DeployHandler - Handles deploying stacks in the corrcet order
func DeployHandler(runstacks map[string]string, stacks map[string]*Stack) {
	// status -  pending, failed, completed
	var status = make(map[string]string)

	for _, stk := range stacks {

		if _, ok := runstacks[stk.Name]; !ok && len(runstacks) > 0 {
			continue
		}

		// Set deploy status & Check if stack exists
		if stk.StackExists() {
			if err := stk.cleanup(); err != nil {
				Log.Error(fmt.Sprintf("Failed to remove stack: [%s] - %s", stk.Name, err.Error()))
				updateState(status, stk.Name, state.failed)
			}

			if stk.StackExists() {
				Log.Info(fmt.Sprintf("stack [%s] already exists...\n", stk.Name))
				continue
			}

		}
		updateState(status, stk.Name, state.pending)

		if len(stk.DependsOn) == 0 {
			wg.Add(1)
			go func(s *Stack) {
				defer wg.Done()

				// Deploy 0 Depency Stacks first - each on their on go routine
				Log.Info(fmt.Sprintf("deploying a template for [%s]", s.Name))

				if err := s.Deploy(); err != nil {
					Log.Error(err.Error())
				}

				updateState(status, s.Name, state.complete)

				// TODO: add deploy Logic here
				return
			}(stk)
			continue
		}

		wg.Add(1)
		go func(s *Stack) {
			Log.Info(fmt.Sprintf("[%s] depends on: %s", s.Name, s.DependsOn))
			defer wg.Done()

			Log.Debug(fmt.Sprintf("Beginning Wait State for Depencies of [%s]"+"\n", s.Name))
			for {
				depts := []string{}
				for _, dept := range s.DependsOn {
					// Dependency wait
					dp, ok := stacks[dept]
					if !ok {
						Log.Error(fmt.Sprintf("Bad dependency: [%s]", dept))
						return
					}

					chk, _ := dp.State()

					switch chk {
					case state.failed:
						updateState(status, dp.Name, state.failed)
					case state.complete:
						updateState(status, dp.Name, state.complete)
					default:
						updateState(status, dp.Name, state.pending)
					}

					mutex.Lock()
					depts = append(depts, status[dept])
					mutex.Unlock()
				}

				if utils.All(depts, state.complete) {
					// Deploy stack once dependencies clear
					Log.Info(fmt.Sprintf("Deploying a template for [%s]", s.Name))

					if err := s.Deploy(); err != nil {
						Log.Error(err.Error())
					}
					return
				}

				for _, v := range depts {
					if v == state.failed {
						Log.Warn(fmt.Sprintf("Deploy Cancelled for stack [%s] due to dependency failure!", s.Name))
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
func TerminateHandler(runstacks map[string]string, stacks map[string]*Stack) {
	for _, stk := range stacks {
		if _, ok := runstacks[stk.Name]; !ok && len(runstacks) > 0 {
			Log.Debug(fmt.Sprintf("%s: not in run.stacks, skipping", stk.Name))
			continue // only process items in the run.stacks unless empty
		}

		// if len(stk.DependsOn) == 0 {
		wg.Add(1)
		go func(s *Stack) {
			defer wg.Done()

			// create ticker
			tick := time.NewTicker(time.Millisecond * 1500)
			defer tick.Stop()

			// Reverse depency look-up so termination waits for all stacks
			// which depend on it, to finish terminating first.
			for {
				for _, stk := range stacks {
					if utils.StringIn(s.Name, stk.DependsOn) {
						Log.Info(fmt.Sprintf("[%s]: Depends on [%s].. Waiting for dependency to terminate", stk.Name, s.Name))
						for _ = range tick.C {
							if !stk.StackExists() {
								break
							}
						}
					}
				}

				s.terminate()
				return
			}

		}(stk)
	}

	// Wait for go routines to complete
	wg.Wait()
}
