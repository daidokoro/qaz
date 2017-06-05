package stacks

import "fmt"

// all - returns true if all items in array the same as the given string
func all(a []string, s string) bool {
	for _, str := range a {
		if s != str {
			return false
		}
	}
	return true
}

// stringIn - returns true if string in array
func stringIn(s string, a []string) bool {
	Log.Debug(fmt.Sprintf("Checking If [%s] is in: %s", s, a))
	for _, str := range a {
		if str == s {
			return true
		}
	}
	return false
}

// cleanup functions in create_failed or delete_failed states
func (s *Stack) cleanup() error {
	Log.Info(fmt.Sprintf("Running stack cleanup on [%s]", s.Name))
	resp, err := s.state()
	if err != nil {
		return err
	}

	if resp == state.failed {
		if err := s.terminate(); err != nil {
			return err
		}
	}
	return nil
}
