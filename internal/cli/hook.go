package cli

import "errors"

// runHook: lightweight target for Consumer lifecycle hooks.
// Reads event JSON from stdin, delegates to capture. Must stay fast (ms budget).
func runHook(args []string) error {
	return errors.New("not implemented: pending RFC-0003 §6 TBD")
}
