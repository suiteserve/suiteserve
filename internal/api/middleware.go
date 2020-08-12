package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func newLogMiddleware(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("<%s> %s %s", r.RemoteAddr, r.Method, r.URL)
		h.ServeHTTP(w, r)
	}
}

func newGetHeadMiddleware(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
				http.StatusMethodNotAllowed)
			return
		}
		h.ServeHTTP(w, r)
	}
}

func newSecMiddleware(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("strict-transport-security", "max-age=31536000")
		h.ServeHTTP(w, r)
	}
}

func newUiSecMiddleware(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

type FileMeta interface {
	Name() string
	ContentType() string
}

type FileMetaRepo interface {
	FileMeta(ctx context.Context, id string) (FileMeta, error)
}

func newUserContentMiddleware(repo FileMetaRepo,
	h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/")
		meta, err := repo.FileMeta(r.Context(), id)
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
			fmt.Sprintf("attachment; filename=%q", meta.Name()))
		w.Header().Set("content-security-policy",
			"sandbox; default-src 'none';")
		w.Header().Set("content-type", meta.ContentType())
		w.Header().Set("x-content-type-options", "nosniff")
		h.ServeHTTP(w, r)
	}
}

func isNotFound(err error) bool {
	var foundErr interface {
		Found() bool
	}
	return errors.As(err, &foundErr) && !foundErr.Found()
}
