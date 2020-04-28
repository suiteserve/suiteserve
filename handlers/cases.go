package handlers

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/tmazeika/testpass/database"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
)

func (s *srv) caseHandler(res http.ResponseWriter, req *http.Request) {
	caseId, ok := mux.Vars(req)["case_id"]
	if !ok {
		log.Panicln("req param 'case_id' not found")
	}

	switch req.Method {
	case http.MethodGet:
		s.getCaseHandler(res, req, caseId)
	case http.MethodPatch:
		s.patchCaseHandler(res, req, caseId)
	case http.MethodDelete:
		s.deleteCaseHandler(res, req, caseId)
	default:
		log.Panicf("method '%s' not implemented\n", req.Method)
	}
}

func (s *srv) getCaseHandler(res http.ResponseWriter, req *http.Request, caseId string) {
	caseRun, err := s.db.WithContext(req.Context()).CaseRun(caseId)
	if errors.Is(err, database.ErrNotFound) {
		httpError(res, errNotFound, http.StatusNotFound)
	} else if err != nil {
		log.Printf("get case run: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
	} else {
		httpJson(res, caseRun, http.StatusOK)
	}
}

func (s *srv) patchCaseHandler(res http.ResponseWriter, req *http.Request, caseId string) {
	var caseRun database.UpdateCaseRun
	if err := json.NewDecoder(req.Body).Decode(&caseRun); err != nil {
		httpError(res, errBadJson, http.StatusBadRequest)
		return
	}
	err := s.db.WithContext(req.Context()).UpdateCaseRun(caseId, caseRun)
	if errors.Is(err, database.ErrInvalidModel) {
		httpError(res, errBadJson, http.StatusBadRequest)
	} else if err != nil {
		log.Printf("update case run: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
	} else {
		res.WriteHeader(http.StatusNoContent)
	}
}

func (s *srv) deleteCaseHandler(res http.ResponseWriter, req *http.Request, caseId string) {
	err := s.db.WithContext(req.Context()).DeleteCaseRun(caseId)
	if errors.Is(err, database.ErrNotFound) {
		httpError(res, errNotFound, http.StatusNotFound)
	} else if err != nil {
		log.Printf("delete case run: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
	} else {
		res.WriteHeader(http.StatusNoContent)
	}
}

func (s *srv) caseCollectionHandler(res http.ResponseWriter, req *http.Request) {
	suiteId, ok := mux.Vars(req)["suite_id"]
	if !ok {
		log.Panicln("req param 'suite_id' not found")
	}

	switch req.Method {
	case http.MethodGet:
		s.getCaseCollectionHandler(res, req, suiteId)
	case http.MethodPost:
		s.postCaseCollectionHandler(res, req, suiteId)
	default:
		log.Panicf("method '%s' not implemented\n", req.Method)
	}
}

func (s *srv) getCaseCollectionHandler(res http.ResponseWriter, req *http.Request, suiteId string) {
	formValNum := req.FormValue("num")
	var caseNum *uint
	if formValNum != "" {
		if num, err := parseUint(formValNum); err != nil {
			httpError(res, errBadQuery, http.StatusBadRequest)
			return
		} else {
			caseNum = &num
		}
	}

	caseRuns, err := s.db.WithContext(req.Context()).AllCaseRuns(suiteId, caseNum)
	if err != nil {
		log.Printf("get all case runs for suite run: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
	} else {
		httpJson(res, caseRuns, http.StatusOK)
	}
}

func (s *srv) postCaseCollectionHandler(res http.ResponseWriter, req *http.Request, suiteId string) {
	var caseRun database.NewCaseRun
	if err := json.NewDecoder(req.Body).Decode(&caseRun); err != nil {
		httpError(res, errBadJson, http.StatusBadRequest)
		return
	}
	id, err := s.db.WithContext(req.Context()).NewCaseRun(suiteId, caseRun)
	if errors.Is(err, database.ErrInvalidModel) {
		httpError(res, errBadJson, http.StatusBadRequest)
		return
	} else if err != nil {
		log.Printf("create case run: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
		return
	}

	loc, err := s.router.Get("case").URL(
		"suite_id", suiteId,
		"case_id", id)
	if err != nil {
		log.Printf("build case url: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
		return
	}

	res.Header().Set("Location", loc.String())
	httpJson(res, bson.M{"id": id}, http.StatusCreated)
}
