package repo

import "fmt"

type errBadMask struct {
	cause error
}

func (e errBadMask) Error() string {
	return fmt.Sprintf("bad mask: %v", e.cause)
}

func (e errBadMask) BadMask() bool {
	return true
}

func (e errBadMask) Unwrap() error {
	return e.cause
}

type errNotFound struct{}

func (e errNotFound) Error() string {
	return "not found"
}

func (e errNotFound) NotFound() bool {
	return true
}
