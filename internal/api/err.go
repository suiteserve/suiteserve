package api

import (
	"errors"
	"log"
	"net/http"
)

type errHttp struct {
	error string
	code  int
	cause error
}

func (e errHttp) Error() string {
	if e.error == "" {
		return http.StatusText(e.code)
	}
	return e.error
}

func (e errHttp) Unwrap() error {
	return e.cause
}

type errHandlerFunc func(w http.ResponseWriter, r *http.Request) error

func (f errHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := f(w, r); err != nil {
		herr := errHttp{code: http.StatusInternalServerError, cause: err}
		if !errors.As(err, &herr) && isNotFound(err) {
			herr.code = http.StatusNotFound
		}
		text := herr.Error()
		if herr.cause != nil {
			text += ": " + herr.cause.Error()
		}
		log.Printf("<%s> %d %s", r.RemoteAddr, herr.code, text)
		http.Error(w, herr.Error(), herr.code)
	}
}

func isNotFound(err error) bool {
	var errNotFound interface {
		NotFound()
	}
	return errors.As(err, &errNotFound)
}

func isBadInput(err error) bool {
	var errBadInput interface {
		BadInput()
	}
	return errors.As(err, &errBadInput)
}
