package repo

import (
	"fmt"
	"github.com/asdine/storm/v3"
)

type errBadJson struct{
	cause error
}

func (e errBadJson) Error() string {
	return fmt.Sprintf("bad json: %v", e.cause)
}

func (errBadJson) BadJson() bool {
	return true
}

func (e errBadJson) Unwrap() error {
	return e.cause
}

type errNotFound struct{}

func (errNotFound) Error() string {
	return "not found"
}

func (errNotFound) NotFound() bool {
	return true
}

func wrapNotFoundErr(err error) error {
	if err == storm.ErrNotFound {
		return errNotFound{}
	}
	return err
}
