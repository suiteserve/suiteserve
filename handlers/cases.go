package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tmazeika/testpass/database"
	"net/http"
)

func (s *srv) caseHandler(res http.ResponseWriter, req *http.Request) error {
	return oneArgHandlerMap{
		http.MethodGet:   s.getCaseHandler,
		http.MethodPatch: s.patchCaseHandler,
	}.handle(res, req, "case_id")
}

func (s *srv) getCaseHandler(res http.ResponseWriter, req *http.Request, id string) error {
	_case, err := s.db.WithContext(req.Context()).Case(id)
	if errors.Is(err, database.ErrNotFound) {
		return errNotFound(errors.New(id))
	} else if err != nil {
		return fmt.Errorf("get case run: %v", err)
	}
	return writeJson(res, http.StatusOK, _case)
}

func (s *srv) patchCaseHandler(res http.ResponseWriter, req *http.Request, id string) error {
	var _case database.UpdateCase
	if err := json.NewDecoder(req.Body).Decode(&_case); err != nil {
		return errBadJson(err)
	}

	err := s.db.WithContext(req.Context()).UpdateCase(id, _case)
	if errors.Is(err, database.ErrInvalidModel) {
		return errBadJson(err)
	} else if errors.Is(err, database.ErrNotFound) {
		return errNotFound(errors.New(id))
	} else if err != nil {
		return fmt.Errorf("update case run: %v", err)
	}

	res.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *srv) caseCollectionHandler(res http.ResponseWriter, req *http.Request) error {
	return oneArgHandlerMap{
		http.MethodGet:  s.getCaseCollectionHandler,
		http.MethodPost: s.postCaseCollectionHandler,
	}.handle(res, req, "suite_id")
}

func (s *srv) getCaseCollectionHandler(res http.ResponseWriter, req *http.Request, suiteId string) error {
	num, ok, err := parseUint(req.FormValue("num"))
	if err != nil {
		return errBadQuery(err)
	}
	var numPtr *uint
	if ok {
		numPtr = &num
	}

	cases, err := s.db.WithContext(req.Context()).AllCases(suiteId, numPtr)
	if err != nil {
		return fmt.Errorf("get all cases for suite: %v", err)
	}
	return writeJson(res, http.StatusOK, cases)
}

func (s *srv) postCaseCollectionHandler(res http.ResponseWriter, req *http.Request, suiteId string) error {
	var _case database.NewCase
	if err := json.NewDecoder(req.Body).Decode(&_case); err != nil {
		return errBadJson(err)
	}

	id, err := s.db.WithContext(req.Context()).NewCase(suiteId, _case)
	if errors.Is(err, database.ErrInvalidModel) {
		return errBadJson(err)
	} else if err != nil {
		return fmt.Errorf("new case: %v", err)
	}

	loc, err := s.router.Get("case").URL("case_id", id)
	if err != nil {
		return fmt.Errorf("build case url: %v", err)
	}

	res.Header().Set("Location", loc.String())
	return writeJson(res, http.StatusCreated, map[string]string{"id": id})
}
