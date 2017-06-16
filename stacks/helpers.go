package stacks

import "fmt"

// cleanup functions in create_failed or delete_failed states
func (s *Stack) cleanup() error {
	Log.Debug(fmt.Sprintf("Running stack cleanup on [%s]", s.Name))
	resp, err := s.State()
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
