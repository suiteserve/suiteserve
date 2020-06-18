package rest

import (
	"net/http"
	"os"
	"path/filepath"
)

func newFrontendHandler(publicDir string) http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		path, err := filepath.Abs(r.URL.Path)
		if err != nil {
			return errBadRequest(err)
		}
		path = filepath.Join(publicDir, path)
		_, err = os.Stat(path)
		if os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(publicDir, "index.html"))
			return nil
		} else if err != nil {
			return err
		}
		http.FileServer(http.Dir(publicDir)).ServeHTTP(w, r)
		return nil
	})
}
