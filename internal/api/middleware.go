package api

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func logMw(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("<%s> %s %s", r.RemoteAddr, r.Method, r.URL)
		h.ServeHTTP(w, r)
	}
}

func methodsMw(methods ...string) func(h http.Handler) http.HandlerFunc {
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

func secMw(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("strict-transport-security", "max-age=31536000")
		h.ServeHTTP(w, r)
	}
}

func uiHandler(dir string) errHandlerFunc {
	fs := http.FileServer(http.Dir(dir))
	return func(w http.ResponseWriter, r *http.Request) error {
		path, err := filepath.Abs(r.URL.Path)
		if err != nil {
			return errHttp{code: http.StatusBadRequest, cause: err}
		}
		path = filepath.Join(dir, path)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
		} else if err != nil {
			return err
		} else {
			fs.ServeHTTP(w, r)
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

func userContentHandler(mr FileMetaRepo, dir string) errHandlerFunc {
	fs := http.FileServer(http.Dir(dir))
	return func(w http.ResponseWriter, r *http.Request) error {
		id := strings.TrimPrefix(r.URL.Path, "/")
		m, err := mr.FileMeta(id)
		if err != nil {
			return err
		}
		w.Header().Set("content-disposition",
			fmt.Sprintf("attachment; filename=%q", m.Name()))
		w.Header().Set("content-security-policy",
			"sandbox; default-src 'none';")
		w.Header().Set("content-type", m.ContentType())
		w.Header().Set("x-content-type-options", "nosniff")
		fs.ServeHTTP(w, r)
		return nil
	}
}

func getVar(r *http.Request, k string) string {
	v, ok := mux.Vars(r)[k]
	if !ok {
		panic(fmt.Sprintf("var %q not found", k))
	}
	return v
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
