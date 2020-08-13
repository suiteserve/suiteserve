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

// func ParamMap(params map[string]http.Handler) http.HandlerFunc {
// 	type param struct {
// 		root bool
// 		parts []string
// 		h http.Handler
// 	}
// 	paramsCpy := make([]param, len(params))
// 	for pattern, h := range params {
// 		paramsCpy = append(paramsCpy, param{
// 			root: strings.HasSuffix(pattern, "/"),
// 			parts: strings.Split(pattern, "/"),
// 			h: h,
// 		})
// 	}
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		parts := strings.Split(r.URL.Path, "/")
// 		for _, p := range paramsCpy {
//
//
// 			matches := pattern.FindStringSubmatch(r.URL.Path)
// 			if len(matches) == 0 {
// 				continue
// 			}
// 			names := pattern.SubexpNames()[1:]
// 			m := make(map[string]string, len(matches[1:]))
// 			for i, v := range matches[1:] {
// 				if v != "" {
// 					m[names[i]] = v
// 				}
// 			}
// 			ctx := context.WithValue(r.Context(), paramKey, m)
// 			h.ServeHTTP(w, r.WithContext(ctx))
// 			return
// 		}
// 		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
// 	}
// }

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
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
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
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
				http.StatusMethodNotAllowed)
		}
	}
}

func NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}
