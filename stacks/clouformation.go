package stacks

import (
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/daidokoro/qaz/log"
	"github.com/daidokoro/qaz/utils"
)

// stat type for handling all stack states
type status struct {
	sync.Mutex
	cache    map[string]string
	pending  string
	failed   string
	complete string
}

func (s *status) update(name, ste string) {
	s.Lock()
	defer s.Unlock()
	log.Debug("Updating Stack Status Map: %s - %s", name, ste)
	s.cache[name] = ste
	return
}

func (s *status) get(name string) string {
	s.Lock()
	defer s.Unlock()
	if v, ok := s.cache[name]; ok {
		return v
	}
	return ""
}

var state = status{
	cache:    make(map[string]string),
	complete: "complete",
	pending:  "pending",
	failed:   "failed",
}

// Exports - prints all cloudformation exports
func Exports(session *session.Session) error {

	svc := cloudformation.New(session)

	exportParams := &cloudformation.ListExportsInput{}

	log.Debug("calling [ListExports] with parameters: %s", exportParams)
	exports, err := svc.ListExports(exportParams)

	if err != nil {
		return err
	}

	for _, i := range exports.Exports {

		fmt.Printf("Export Name: %s\nExport Value: %s\n--\n", log.ColorString(*i.Name, log.MAGENTA), *i.Value)
	}

	return nil
}

// DeployHandler - Handles deploying stacks in the corrcet order
// TODO - this is still a fairly convoluted function... need to splify
func DeployHandler(m *Map) {
	var wg sync.WaitGroup
	// kick off tail mechanism
	tail = make(chan *TailServiceInput)
	go TailService(tail)

	m.Range(func(k string, s *Stack) bool {
		if !s.Actioned {
			return true
		}

		// Set deploy status & Check if stack exists
		if s.StackExists() {
			if err := s.cleanup(); err != nil {
				log.Error("failed to remove stack: [%s] - %v", s.Name, err)
				// updateState(status, stk.Name, state.failed)
				state.update(s.Name, state.failed)
			}

			if s.StackExists() {
				log.Info("stack [%s] already exists...\n", s.Name)
				return true
			}

		}
		state.update(s.Name, state.pending)

		if len(s.DependsOn) == 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()

				// Deploy 0 Depency Stacks first - each on their on go routine
				log.Info("deploying a template for [%s]", s.Name)

				if err := s.Deploy(); err != nil {
					log.Error(err.Error())
				}

				state.update(s.Name, state.complete)

				// TODO: add deploy logic here
				return
			}()
			return true
		}

		wg.Add(1)
		go func() {
			log.Info("[%s] depends on: %s", s.Name, s.DependsOn)
			defer wg.Done()

			log.Debug("Beginning Wait State for Depencies of [%s]"+"\n", s.Name)
			for {
				depts := []string{}
				for _, dept := range s.DependsOn {
					// Dependency wait
					dp, ok := m.Get(dept)
					if !ok {
						log.Error("Bad dependency: [%s]", dept)
						return
					}

					chk, _ := dp.State()

					switch chk {
					case state.failed:
						state.update(dp.Name, state.failed)
					case state.complete:
						state.update(dp.Name, state.complete)
					default:
						state.update(dp.Name, state.pending)
					}

					depts = append(depts, state.get(dept))

				}

				if utils.All(depts, state.complete) {
					// Deploy stack once dependencies clear
					log.Info("deploying a template for [%s]", s.Name)

					if err := s.Deploy(); err != nil {
						log.Error(err.Error())
					}
					return
				}

				for _, v := range depts {
					if v == state.failed {
						log.Warn("deploy Cancelled for stack [%s] due to dependency failure!", s.Name)
						return
					}
				}

				time.Sleep(time.Second * 1)
			}
		}()
		return true

	})

	// Wait for go routines to complete
	wg.Wait()
}

// TerminateHandler - Handles terminating stacks in the correct order
func TerminateHandler(m *Map) {
	// kick off tail mechanism
	tail = make(chan *TailServiceInput)
	var wg sync.WaitGroup
	go TailService(tail)

	m.Range(func(k string, s *Stack) bool {
		if !s.Actioned {
			log.Debug("%s: not actioned, skipping", s.Name)
			return true
		}

		// if len(stk.DependsOn) == 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// create ticker
			tick := time.NewTicker(time.Millisecond * 1500)
			defer tick.Stop()

			// Reverse depency look-up so termination waits for all stacks
			// which depend on it, to finish terminating first.
			for {
				m.Range(func(k string, stk *Stack) bool {
					if utils.StringIn(s.Name, stk.DependsOn) {
						log.Info("[%s]: Depends on [%s].. Waiting for dependency to terminate", stk.Name, s.Name)
						for range tick.C {
							if !stk.StackExists() {
								break
							}
						}
					}
					return true
				})

				if err := s.terminate(); err != nil {
					log.Error("error deleting stack: [%s] - %v", s.Name, err)
				}
				return
			}

		}()
		return true
	})

	// Wait for go routines to complete
	wg.Wait()
}
