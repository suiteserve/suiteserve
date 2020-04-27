package handlers

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/tmazeika/testpass/database"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"strconv"
)

func (s *srv) caseHandler(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	suiteId, ok := params["suite_id"]
	if !ok {
		log.Panicln("req param 'suite_id' not found")
	}
	caseNumStr, ok := params["case_num"]
	if !ok {
		log.Panicln("req param 'case_num' not found")
	}
	caseNum, err := strconv.ParseUint(caseNumStr, 10, 64)
	if err != nil {
		log.Panicln("req param 'case_num' NaN")
	}

	switch req.Method {
	case http.MethodGet:
		s.getCaseHandler(res, req, suiteId, uint(caseNum))
	case http.MethodPatch:
		s.patchCaseHandler(res, req, suiteId, uint(caseNum))
	case http.MethodDelete:
		s.deleteCaseHandler(res, req, suiteId, uint(caseNum))
	default:
		log.Panicf("method '%s' not implemented\n", req.Method)
	}
}

func (s *srv) getCaseHandler(res http.ResponseWriter, req *http.Request, suiteId string, caseNum uint) {
	caseRuns, err := s.db.WithContext(req.Context()).CaseRuns(suiteId, caseNum)
	if errors.Is(err, database.ErrNotFound) {
		httpError(res, errNotFound, http.StatusNotFound)
	} else if err != nil {
		log.Printf("get case runs: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
	} else {
		httpJson(res, caseRuns, http.StatusOK)
	}
}

func (s *srv) patchCaseHandler(res http.ResponseWriter, req *http.Request, suiteId string, caseNum uint) {
	var caseRun database.UpdateCaseRun
	if err := json.NewDecoder(req.Body).Decode(&caseRun); err != nil {
		httpError(res, errBadJson, http.StatusBadRequest)
		return
	}
	err := s.db.WithContext(req.Context()).UpdateCaseRun(suiteId, caseNum, caseRun)
	if errors.Is(err, database.ErrInvalidModel) {
		httpError(res, errBadJson, http.StatusBadRequest)
	} else if err != nil {
		log.Printf("update case run: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
	} else {
		res.WriteHeader(http.StatusNoContent)
	}
}

func (s *srv) deleteCaseHandler(res http.ResponseWriter, req *http.Request, suiteId string, caseNum uint) {
	err := s.db.WithContext(req.Context()).DeleteCaseRuns(suiteId, caseNum)
	if errors.Is(err, database.ErrNotFound) {
		httpError(res, errNotFound, http.StatusNotFound)
	} else if err != nil {
		log.Printf("delete case runs: %v\n", err)
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
	caseRuns, err := s.db.WithContext(req.Context()).AllCaseRuns(suiteId)
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
		"case_num", strconv.FormatUint(uint64(caseRun.Num), 10))
	if err != nil {
		log.Printf("build case url: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
		return
	}

	res.Header().Set("Location", loc.String())
	httpJson(res, bson.M{"id": id}, http.StatusCreated)
}
