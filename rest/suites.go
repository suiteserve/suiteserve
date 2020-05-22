package rest

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *srv) getSuiteHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		suite, err := s.repos.Suites().Find(id)
		if err != nil {
			return fmt.Errorf("get suite: %v", err)
		}
		return writeJson(w, http.StatusOK, suite)
	})
}

func (s *srv) deleteSuiteHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		err := s.repos.Suites().Delete(id)
		if err != nil {
			return fmt.Errorf("delete suite: %v", err)
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}

func (s *srv) getSuiteCollectionHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		afterId := parseString(r.FormValue("after_id"))
		limit, err := parseInt64(r.FormValue("limit"))
		if err != nil {
			return errBadQuery(err)
		} else if limit != nil && *limit < 1 {
			return errBadQuery(errors.New("limit must be positive"))
		}
		if limit == nil {
			l := int64(10)
			limit = &l
		}

		suites, err := s.repos.Suites().Page(afterId, *limit, false)
		if err != nil {
			return fmt.Errorf("get all suites: %v", err)
		}
		return writeJson(w, http.StatusOK, suites)
	})
}

func (s *srv) deleteSuiteCollectionHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		err := s.repos.Suites().DeleteAll()
		if err != nil {
			return fmt.Errorf("delete all suites: %v", err)
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}
