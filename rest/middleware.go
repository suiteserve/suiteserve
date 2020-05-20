package rest

import (
	"log"
	"net/http"
	"strings"
)

func methodOverrideMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			m := strings.ToUpper(r.FormValue("_method"))
			if m == http.MethodPut || m == http.MethodPatch || m == http.MethodDelete {
				r.Method = m
			}
		}
		h.ServeHTTP(w, r)
	})
}

func loggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL.String())
		h.ServeHTTP(w, r)
	})
}

func defaultSecureHeadersMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-security-policy", "sandbox; "+
			"default-src 'none'; "+
			"base-uri 'none'; "+
			"form-action 'none'; "+
			"frame-ancestors 'none'; "+
			"img-src 'self';")
		w.Header().Set("strict-transport-security", "max-age=31536000")
		w.Header().Set("x-content-type-options", "nosniff")
		w.Header().Set("x-frame-options", "deny")
		h.ServeHTTP(w, r)
	})
}

func frontendSecureHeadersMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-security-policy", "block-all-mixed-content; "+
			"default-src 'none'; "+
			"base-uri 'none'; "+
			"connect-src 'self'; "+
			"font-src https://fonts.gstatic.com; "+
			"form-action 'self'; "+
			"frame-ancestors 'none'; "+
			"img-src 'self'; "+
			"script-src 'self' 'unsafe-eval'; "+
			"style-src 'self' https://fonts.googleapis.com;")
		h.ServeHTTP(w, r)
	})
}
