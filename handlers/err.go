package handlers

import (
	"fmt"
	"log"
	"net/http"
)

type restError struct {
	code int
	name string
}

func (e restError) Error() string {
	return fmt.Sprintf("%d: %s", e.code, e.name)
}

var (
	errBadFile  = restError{http.StatusBadRequest, "bad_file"}
	errBadJson  = restError{http.StatusBadRequest, "bad_json"}
	errBadQuery = restError{http.StatusBadRequest, "bad_query"}
	errNotFound = restError{http.StatusNotFound, "not_found"}
)

func errorHandler(h func(res http.ResponseWriter, req *http.Request) error) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		write := func(code int, name string) {
			if err := writeJson(res, code, map[string]string{"error": name}); err != nil {
				log.Printf("%s %v\n", req.RemoteAddr, err)
			}
		}
		if err := h(res, req); err != nil {
			if restErr, ok := err.(restError); ok {
				log.Printf("%s status %v\n", req.RemoteAddr, err)
				write(restErr.code, restErr.name)
			} else {
				log.Printf("%s status 500: %v\n", req.RemoteAddr, err)
				write(http.StatusInternalServerError, "unknown")
			}
		}
	})
}
