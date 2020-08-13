package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

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

func newUserContentMiddleware(repo FileMetaRepo, h http.Handler) http.Handler {
	return errHandler(func(w http.ResponseWriter, r *http.Request) error {
		id := strings.TrimPrefix(r.URL.Path, "/")
		meta, err := repo.FileMeta(r.Context(), id)
		if isNotFound(err) {
			return httpError{code: http.StatusNotFound, cause: err}
		} else if err != nil {
			return err
		}
		w.Header().Set("content-disposition",
			fmt.Sprintf("attachment; filename=%q", meta.Name()))
		w.Header().Set("content-security-policy",
			"sandbox; default-src 'none';")
		w.Header().Set("content-type", meta.ContentType())
		w.Header().Set("x-content-type-options", "nosniff")
		h.ServeHTTP(w, r)
		return nil
	})
}
