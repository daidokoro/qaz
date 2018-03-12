package stacks

import "sync"

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
