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
		httpErr := errHttp{
			code:  http.StatusInternalServerError,
			cause: err,
		}
		if !errors.As(err, &httpErr) && isNotFound(err) {
			httpErr.code = http.StatusNotFound
		}
		text := httpErr.Error()
		if httpErr.cause != nil {
			text += ": " + httpErr.cause.Error()
		}
		log.Printf("<%s> %d %s", r.RemoteAddr, httpErr.code, text)
		http.Error(w, httpErr.Error(), httpErr.code)
	}
}

func isNotFound(err error) bool {
	var errNotFound interface {
		NotFound() bool
	}
	return errors.As(err, &errNotFound)
}
