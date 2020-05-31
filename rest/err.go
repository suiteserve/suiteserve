package rest

import (
	"fmt"
	"log"
	"net/http"
)

type httpError struct {
	code  int
	name  string
	cause error
}

func (e httpError) Error() string {
	return fmt.Sprintf("%d %s: %v", e.code, e.name, e.cause)
}

func errBadQuery(cause error) error {
	return httpError{http.StatusBadRequest, "bad_query", cause}
}

func errNotFound(cause error) error {
	return httpError{http.StatusNotFound, "not_found", cause}
}

func errorHandler(h func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		write := func(code int, name string) {
			err := writeJson(w, code, map[string]interface{}{"error": name})
			if err != nil {
				log.Printf("%s %v\n", r.RemoteAddr, err)
			}
		}
		if err := h(w, r); err != nil {
			if httpErr, ok := err.(httpError); ok {
				log.Printf("%s status %v\n", r.RemoteAddr, err)
				write(httpErr.code, httpErr.name)
			} else {
				log.Printf("%s status 500: %v\n", r.RemoteAddr, err)
				write(http.StatusInternalServerError, "unknown")
			}
		}
	})
}
