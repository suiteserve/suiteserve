package api

import "net/http"

func NewV1() http.Handler {
	var mux http.ServeMux
	return &mux
}
