package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type ctxKey int

const pathParamKey ctxKey = iota

func logMw(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("<%s> %s %s", r.RemoteAddr, r.Method, r.URL)
		h.ServeHTTP(w, r)
	}
}

func pathParamMw(ignorePrefix string, h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, ignorePrefix)
		if strings.ContainsRune(path, '/') {
			notFound()(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), pathParamKey, path)
		h.ServeHTTP(w, r.WithContext(ctx))
	}
}

func pathParam(r *http.Request) string {
	v, ok := r.Context().Value(pathParamKey).(string)
	if !ok {
		panic("path param not found")
	}
	return v
}

func methodsMw(methods ...string) func(h http.Handler) http.HandlerFunc {
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
			methodNotAllowed()(w, r)
		}
	}
}

// func methodMapMw(methods map[string]http.Handler) http.HandlerFunc {
// 	methodsCpy := make(map[string]http.Handler, len(methods))
// 	for m, h := range methods {
// 		methodsCpy[strings.ToUpper(m)] = h
// 	}
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		for m, h := range methodsCpy {
// 			if m == r.Method {
// 				h.ServeHTTP(w, r)
// 				return
// 			}
// 		}
// 		methodNotAllowed()(w, r)
// 	}
// }

func secMw(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("strict-transport-security", "max-age=31536000")
		h.ServeHTTP(w, r)
	}
}

func uiSecMw(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		h.ServeHTTP(w, r)
	}
}

func uiHandler(publicDir string) errHandlerFunc {
	fileRepo := http.FileServer(http.Dir(publicDir))
	return func(w http.ResponseWriter, r *http.Request) error {
		path, err := filepath.Abs(r.URL.Path)
		if err != nil {
			return errHttp{code: http.StatusBadRequest, cause: err}
		}
		path = filepath.Join(publicDir, path)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(publicDir, "index.html"))
		} else if err != nil {
			return err
		} else {
			fileRepo.ServeHTTP(w, r)
		}
		return nil
	}
}

type FileMeta interface {
	Name() string
	ContentType() string
}

type FileMetaRepo interface {
	FileMeta(id string) (FileMeta, error)
}

func userContentHandler(repo FileMetaRepo, dir string) errHandlerFunc {
	fs := http.FileServer(http.Dir(dir))
	return func(w http.ResponseWriter, r *http.Request) error {
		id := strings.TrimPrefix(r.URL.Path, "/")
		meta, err := repo.FileMeta(id)
		if isNotFound(err) {
			return errHttp{code: http.StatusNotFound, cause: err}
		} else if err != nil {
			return err
		}
		w.Header().Set("content-disposition",
			fmt.Sprintf("attachment; filename=%q", meta.Name()))
		w.Header().Set("content-security-policy",
			"sandbox; default-src 'none';")
		w.Header().Set("content-type", meta.ContentType())
		w.Header().Set("x-content-type-options", "nosniff")
		fs.ServeHTTP(w, r)
		return nil
	}
}

func methodNotAllowed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
	}
}

func notFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound),
			http.StatusNotFound)
	}
}
