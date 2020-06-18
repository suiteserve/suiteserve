package rest

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

type httpError struct {
	code  int
	name  string
	cause error
}

func (e *httpError) Unwrap() error {
	return e.cause
}

func (e *httpError) Error() string {
	return fmt.Sprintf("%d %s: %v", e.code, e.name, e.cause)
}

func errBadRequest(cause error) error {
	return &httpError{http.StatusBadRequest, "bad_request", cause}
}

func errNotFound(cause error) error {
	return &httpError{http.StatusNotFound, "not_found", cause}
}

func errorHandler(h func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}
		var httpErr *httpError
		if !errors.As(err, &httpErr) {
			log.Printf("%s status 500: %v\n", r.RemoteAddr, err)
			writeErr(w, r, http.StatusInternalServerError, "unknown")
			return
		}
		log.Printf("%s status %v\n", r.RemoteAddr, httpErr)
		writeErr(w, r, httpErr.code, httpErr.name)
	})
}

func writeErr(w http.ResponseWriter, r *http.Request, code int, name string) {
	err := writeJson(w, code, map[string]interface{}{"error": name})
	if err != nil {
		log.Printf("%s write json: %v\n", r.RemoteAddr, err)
	}
}
