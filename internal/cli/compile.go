package cli

import "errors"

// runCompile: incremental scan (cache cursor) -> append new Facts ->
// replay Trust view -> skeleton -> attach Enhancements (Mode 2/3, RFC-0003 §4 fallback).
func runCompile(args []string) error {
	return errors.New("not implemented")
}
