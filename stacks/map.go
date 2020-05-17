package stacks

import (
	"fmt"
	"strings"
	"sync"

	"text/template"

	"github.com/daidokoro/qaz/log"
	"github.com/daidokoro/qaz/utils"
)

// Map type
type Map struct {
	sync.Mutex
	store map[string]*Stack
}

// Get = retuern stack
func (m *Map) Get(k string) (s *Stack, ok bool) {
	s, ok = m.store[k]
	return
}

// MustGet - assumes stack exists at the given Key
// returns value without validation
func (m *Map) MustGet(k string) (s *Stack) {
	return m.store[k]
}

// Add - add stack
func (m *Map) Add(k string, stk *Stack) {
	m.Lock()
	defer m.Unlock()

	if m.store == nil {
		m.store = make(map[string]*Stack)
	}

	if _, ok := m.store[k]; !ok {
		m.store[k] = stk
	}

	return
}

// Range - iterate over map and execute function
// against each k, v pair.
// iteration ends if function returns false
func (m *Map) Range(f func(string, *Stack) bool) {
	for k, v := range m.store {
		if !f(k, v) {
			return
		}
	}
	return
}

// Count - returns number of stacks
func (m *Map) Count() int {
	return len(m.store)
}

// --- Template Functions for StackMap --- //

// StackOutput - stack output reader for template function
func (m *Map) StackOutput(target string) string {
	log.Debug("Deploy-Time function resolving: %s", target)
	req := strings.Split(target, "::")

	s, ok := m.Get(req[0])
	if !ok {
		utils.HandleError(fmt.Errorf("stack_output errror: stack [%s] not found", req[0]))
	}

	utils.HandleError(s.Outputs())

	for _, i := range s.Output.Stacks {
		for _, o := range i.Outputs {
			if *o.OutputKey == req[1] {
				return *o.OutputValue
			}
		}
	}
	utils.HandleError(fmt.Errorf("Stack Output Not found - Stack:%s | Output:%s", req[0], req[1]))
	return ""
}

// AddMapFuncs - add stack map functions to function map
// Note: this map only exists for deploytime functions
// that require access to stack data at runtime.
func (m *Map) AddMapFuncs(t template.FuncMap) {
	// Fetching stackoutputs
	t["stack_output"] = func(target string) string {
		log.Debug("Deploy-Time function resolving: %s", target)
		req := strings.Split(target, "::")

		s, ok := m.Get(req[0])
		if !ok {
			utils.HandleError(fmt.Errorf("stack_output errror: stack [%s] not found", req[0]))
		}

		utils.HandleError(s.Outputs())

		for _, i := range s.Output.Stacks {
			for _, o := range i.Outputs {
				if *o.OutputKey == req[1] {
					return *o.OutputValue
				}
			}
		}
		utils.HandleError(fmt.Errorf("Stack Output Not found - Stack:%s | Output:%s", req[0], req[1]))
		return ""
	}

	t["stack_output_ext"] = func(target string) string {
		log.Debug("Deploy-Time function resolving: %s", target)
		req := strings.Split(target, "::")
		s := m.MustGet(req[0])
		utils.HandleError(s.Outputs())

		for _, i := range s.Output.Stacks {
			for _, o := range i.Outputs {
				if *o.OutputKey == req[1] {
					return *o.OutputValue
				}
			}
		}

		utils.HandleError(fmt.Errorf("Stack Output Not found - Stack:%s | Output:%s", req[0], req[1]))
		return ""
	}
}
