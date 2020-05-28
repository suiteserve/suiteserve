package rest

import (
	"fmt"
	"github.com/gorilla/mux"
	util "github.com/tmazeika/testpass/internal"
	"github.com/tmazeika/testpass/repo"
	"net/http"
	"strconv"
)

func (s *srv) getSuiteHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		suite, err := s.repos.Suites().Find(r.Context(), id)
		if err == repo.ErrNotFound {
			return errNotFound(err)
		} else if err != nil {
			return fmt.Errorf("get suite: %v", err)
		}
		return writeJson(w, http.StatusOK, suite)
	})
}

func (s *srv) deleteSuiteHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		err := s.repos.Suites().Delete(r.Context(), id, util.NowTimeMillis())
		if err == repo.ErrNotFound {
			return errNotFound(err)
		} else if err != nil {
			return fmt.Errorf("delete suite: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}

func (s *srv) getSuiteCollectionHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		fromId := parseStringPtr(r.FormValue("from_id"))
		limit, err := strconv.ParseInt(r.FormValue("limit"), 10, 64)
		if err != nil || limit < 1 {
			limit = 10
		}

		suites, err := s.repos.Suites().Page(r.Context(), fromId, limit, false)
		if err != nil {
			return fmt.Errorf("get all suites: %v", err)
		}
		return writeJson(w, http.StatusOK, suites)
	})
}

func (s *srv) deleteSuiteCollectionHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		err := s.repos.Suites().DeleteAll(r.Context(), util.NowTimeMillis())
		if err != nil {
			return fmt.Errorf("delete all suites: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}
