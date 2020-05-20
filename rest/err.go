package rest

import (
	"fmt"
	"log"
	"net/http"
)

type restError struct {
	code int
	name string
	cause error
}

func (e restError) Error() string {
	return fmt.Sprintf("%d %s: %v", e.code, e.name, e.cause)
}

func errBadJson(cause error) error {
	return restError{http.StatusBadRequest, "bad_json", cause}
}

func errBadQuery(cause error) error {
	return restError{http.StatusBadRequest, "bad_query", cause}
}

func errNotFound(cause error) error {
	return restError{http.StatusNotFound, "not_found", cause}
}

func errorHandler(h func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		write := func(code int, name string) {
			if err := writeJson(w, code, map[string]interface{}{"error": name}); err != nil {
				log.Printf("%s %v\n", r.RemoteAddr, err)
			}
		}
		if err := h(w, r); err != nil {
			if restErr, ok := err.(restError); ok {
				log.Printf("%s status %v\n", r.RemoteAddr, err)
				write(restErr.code, restErr.name)
			} else {
				log.Printf("%s status 500: %v\n", r.RemoteAddr, err)
				write(http.StatusInternalServerError, "unknown")
			}
		}
	})
}
