package rest

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/suiteserve/suiteserve/repo"
	"net/http"
)

type caseFinder interface {
	Case(ctx context.Context, id string) (*repo.Case, error)
	CasesBySuite(ctx context.Context, suiteId string) ([]repo.Case, error)
}

func newGetCaseHandler(finder caseFinder) http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		c, err := finder.Case(r.Context(), id)
		if errors.Is(err, repo.ErrNotFound) {
			return errNotFound(err)
		} else if err != nil {
			return fmt.Errorf("get case: %v", err)
		}
		return writeJson(w, http.StatusOK, c)
	})
}

func newGetCaseCollectionHandler(finder caseFinder) http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		suiteId := mux.Vars(r)["suite_id"]
		cases, err := finder.CasesBySuite(r.Context(), suiteId)
		if errors.Is(err, repo.ErrNotFound) {
			return errNotFound(err)
		} else if err != nil {
			return fmt.Errorf("get cases by suite: %v", err)
		}
		return writeJson(w, http.StatusOK, cases)
	})
}
