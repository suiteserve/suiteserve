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
	type errorRes struct {
		Error string `json:"error"`
	}
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if err := h(res, req); err != nil {
			if restErr, ok := err.(restError); ok {
				log.Printf("%s status %v\n", req.RemoteAddr, err)
				writeJson(res, errorRes{restErr.name}, restErr.code)
			} else {
				log.Printf("%s status 500: %v\n", req.RemoteAddr, err)
				writeJson(res, errorRes{"unknown"}, http.StatusInternalServerError)
			}
		}
	})
}
