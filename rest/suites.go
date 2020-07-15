package rest

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	util "github.com/suiteserve/suiteserve/internal"
	"github.com/suiteserve/suiteserve/repo"
	"net/http"
	"strconv"
)

type suiteFinder interface {
	Suite(ctx context.Context, id string) (*repo.Suite, error)
	SuitePage(ctx context.Context, fromId string, limit int) (*repo.SuitePage, error)
}

type suiteUpdater interface {
	DeleteSuite(ctx context.Context, id string, at int64) error
	DeleteAllSuites(ctx context.Context, at int64) error
}

func newGetSuiteHandler(finder suiteFinder) http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		suite, err := finder.Suite(r.Context(), id)
		if errors.Is(err, repo.ErrNotFound) {
			return errNotFound(err)
		} else if err != nil {
			return fmt.Errorf("get suite: %v", err)
		}
		return writeJson(w, http.StatusOK, suite)
	})
}

func newDeleteSuiteHandler(updater suiteUpdater) http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		err := updater.DeleteSuite(r.Context(), id, util.NowTimeMillis())
		if errors.Is(err, repo.ErrNotFound) {
			return errNotFound(err)
		} else if err != nil {
			return fmt.Errorf("delete suite: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}

func newGetSuiteCollectionHandler(finder suiteFinder) http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		fromId := r.FormValue("from_id")
		limit, err := strconv.ParseInt(r.FormValue("limit"), 10, 32)
		if err != nil || limit < 1 {
			limit = 10
		}

		suites, err := finder.SuitePage(r.Context(), fromId, int(limit))
		if errors.Is(err, repo.ErrNotFound) {
			return errNotFound(err)
		} else if err != nil {
			return fmt.Errorf("get suites page: %v", err)
		}
		return writeJson(w, http.StatusOK, suites)
	})
}

func newDeleteSuiteCollectionHandler(updater suiteUpdater) http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		err := updater.DeleteAllSuites(r.Context(), util.NowTimeMillis())
		if err != nil {
			return fmt.Errorf("delete all suites: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}
