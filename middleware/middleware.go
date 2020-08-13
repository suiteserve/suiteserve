package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
)

type ctxKey int

const paramKey ctxKey = iota

func Log(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("<%s> %s %s", r.RemoteAddr, r.Method, r.URL)
		h.ServeHTTP(w, r)
	}
}

func Param(trimPrefix string, h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, trimPrefix)
		if strings.ContainsRune(path, '/') {
			NotFound().ServeHTTP(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), paramKey, path)
		h.ServeHTTP(w, r.WithContext(ctx))
	}
}

func GetParam(r *http.Request) string {
	v, ok := r.Context().Value(paramKey).(string)
	if !ok {
		panic("param not found")
	}
	return v
}

func MethodMap(methods map[string]http.Handler) http.HandlerFunc {
	methodsCpy := make(map[string]http.Handler, len(methods))
	for m, h := range methods {
		methodsCpy[strings.ToUpper(m)] = h
	}
	return func(w http.ResponseWriter, r *http.Request) {
		for m, h := range methodsCpy {
			if m == r.Method {
				h.ServeHTTP(w, r)
				return
			}
		}
		MethodNotAllowed().ServeHTTP(w, r)
	}
}

func Methods(methods ...string) func(h http.Handler) http.HandlerFunc {
	for i, m := range methods {
		methods[i] = strings.ToUpper(m)
	}
	return func(h http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			for _, m := range methods {
				if m == r.Method {
					h.ServeHTTP(w, r)
					return
				}
			}
			MethodNotAllowed().ServeHTTP(w, r)
		}
	}
}

func MethodNotAllowed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
	}
}

func NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound),
			http.StatusNotFound)
	}
}
