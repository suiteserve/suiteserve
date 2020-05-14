package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tmazeika/testpass/database"
	"net/http"
)

func (s *srv) suiteHandler(res http.ResponseWriter, req *http.Request) error {
	return oneArgHandlerMap{
		http.MethodGet:    s.getSuiteHandler,
		http.MethodPatch:  s.patchSuiteHandler,
		http.MethodDelete: s.deleteSuiteHandler,
	}.handle(res, req, "suite_id")
}

func (s *srv) getSuiteHandler(res http.ResponseWriter, req *http.Request, id string) error {
	suite, err := s.db.WithContext(req.Context()).Suite(id)
	if errors.Is(err, database.ErrNotFound) {
		return errNotFound(errors.New(id))
	} else if err != nil {
		return fmt.Errorf("get suite: %v", err)
	}
	return writeJson(res, http.StatusOK, suite)
}

func (s *srv) patchSuiteHandler(res http.ResponseWriter, req *http.Request, id string) error {
	var suite database.UpdateSuite
	if err := json.NewDecoder(req.Body).Decode(&suite); err != nil {
		return errBadJson(err)
	}

	err := s.db.WithContext(req.Context()).UpdateSuite(id, suite)
	if errors.Is(err, database.ErrInvalidModel) {
		return errBadJson(err)
	} else if errors.Is(err, database.ErrNotFound) {
		return errNotFound(errors.New(id))
	} else if err != nil {
		return fmt.Errorf("update suite: %v", err)
	}

	res.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *srv) deleteSuiteHandler(res http.ResponseWriter, req *http.Request, id string) error {
	_, err := s.db.WithContext(req.Context()).DeleteSuite(id)
	if err != nil {
		return fmt.Errorf("delete suite: %v", err)
	}

	res.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *srv) suiteCollectionHandler(res http.ResponseWriter, req *http.Request) error {
	return noArgHandlerMap{
		http.MethodGet:    s.getSuiteCollectionHandler,
		http.MethodPost:   s.postSuiteCollectionHandler,
		http.MethodDelete: s.deleteSuiteCollectionHandler,
	}.handle(res, req)
}

func (s *srv) getSuiteCollectionHandler(res http.ResponseWriter, req *http.Request) error {
	afterId := parseString(req.FormValue("after_id"))
	limit, err := parseInt64(req.FormValue("limit"))
	if err != nil {
		return errBadQuery(err)
	} else if limit != nil && *limit < 1 {
		return errBadQuery(errors.New("limit must be positive"))
	}

	suites, err := s.db.WithContext(req.Context()).AllSuites(afterId, limit)
	if err != nil {
		return fmt.Errorf("get all suites: %v", err)
	}
	return writeJson(res, http.StatusOK, suites)
}

func (s *srv) postSuiteCollectionHandler(res http.ResponseWriter, req *http.Request) error {
	var suite database.NewSuite
	if err := json.NewDecoder(req.Body).Decode(&suite); err != nil {
		return errBadJson(err)
	}

	id, err := s.db.WithContext(req.Context()).NewSuite(suite)
	if errors.Is(err, database.ErrInvalidModel) {
		return errBadJson(err)
	} else if err != nil {
		return fmt.Errorf("new suite run: %v", err)
	}

	loc, err := s.router.Get("suite").URL("suite_id", id)
	if err != nil {
		return fmt.Errorf("build suite url: %v", err)
	}

	res.Header().Set("Location", loc.String())
	return writeJson(res, http.StatusCreated, map[string]string{"id": id})
}

func (s *srv) deleteSuiteCollectionHandler(res http.ResponseWriter, req *http.Request) error {
	err := s.db.WithContext(req.Context()).DeleteAllSuites()
	if err != nil {
		return fmt.Errorf("delete all suites: %v", err)
	}

	res.WriteHeader(http.StatusNoContent)
	return nil
}
