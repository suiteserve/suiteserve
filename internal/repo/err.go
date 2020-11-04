package repo

import "fmt"

type errNotFound struct{}

func (errNotFound) Error() string {
	return "not found"
}

func (errNotFound) NotFound() {}

type errBadFormat struct {
	error
}

func (e errBadFormat) Error() string {
	return fmt.Sprintf("bad format: %v", e.error)
}

func (e errBadFormat) Unwrap() error {
	return e.error
}

func (e errBadFormat) BadFormat() {}

func errBadId(err error) error {
	return errBadFormat{fmt.Errorf("bad id: %v", err)}
}
