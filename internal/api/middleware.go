package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
)

func newLoggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("<%s> %s %s", r.RemoteAddr, r.Method, r.URL.String())
		h.ServeHTTP(w, r)
	})
}

func newGetMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
				http.StatusMethodNotAllowed)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func newHijackMiddleware(wg *sync.WaitGroup, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := w.(http.Hijacker); !ok {
			log.Printf("cannot hijack resp")
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
		wg.Add(1)
		defer wg.Done()
		h.ServeHTTP(w, r)
	})
}

func newSecurityMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("strict-transport-security", "max-age=31536000")
		h.ServeHTTP(w, r)
	})
}

func newFrontendSecurityMiddleware(h http.Handler) http.Handler {
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

func newUserContentSecurityMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-security-policy", "sandbox; default-src 'none';")
		w.Header().Set("x-content-type-options", "nosniff")
		h.ServeHTTP(w, r)
	})
}

type FileMeta struct {
	Filename    string
	ContentType string
}

type UserContentMetaRepo interface {
	UserContentMeta(ctx context.Context, id string) (*FileMeta, error)
}

func newUserContentMiddleware(repo UserContentMetaRepo, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/")
		meta, err := repo.UserContentMeta(r.Context(), id)
		if isNotFound(err) {
			http.Error(w, http.StatusText(http.StatusNotFound),
				http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
		w.Header().Set("content-disposition",
			fmt.Sprintf("attachment; filename=%q", meta.Filename))
		w.Header().Set("content-type", meta.ContentType)
		h.ServeHTTP(w, r)
	})
}

func isNotFound(err error) bool {
	var foundErr interface {
		Found() bool
	}
	if !errors.As(err, &foundErr) {
		return false
	}
	return !foundErr.Found()
}
