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

func (s *srv) suiteHandler(res http.ResponseWriter, req *http.Request) {
	id, ok := mux.Vars(req)["suite_id"]
	if !ok {
		log.Panicln("req param 'suite_id' not found")
	}

	switch req.Method {
	case http.MethodGet:
		s.getSuiteHandler(res, req, id)
	case http.MethodDelete:
		s.deleteSuiteHandler(res, req, id)
	default:
		log.Panicf("method '%s' not implemented\n", req.Method)
	}
}

func (s *srv) getSuiteHandler(res http.ResponseWriter, req *http.Request, id string) {
	suiteRun, err := s.db.WithContext(req.Context()).SuiteRun(id)
	if errors.Is(err, database.ErrNotFound) {
		httpError(res, errNotFound, http.StatusNotFound)
	} else if err != nil {
		log.Printf("get suite run: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
	} else {
		httpJson(res, suiteRun, http.StatusOK)
	}
}

func (s *srv) deleteSuiteHandler(res http.ResponseWriter, req *http.Request, id string) {
	err := s.db.WithContext(req.Context()).DeleteSuiteRun(id)
	if errors.Is(err, database.ErrNotFound) {
		httpError(res, errNotFound, http.StatusNotFound)
	} else if err != nil {
		log.Printf("delete suite run: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
	} else {
		res.WriteHeader(http.StatusNoContent)
	}
}

func (s *srv) suiteCollectionHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		s.getSuiteCollectionHandler(res, req)
	case http.MethodPost:
		s.postSuiteCollectionHandler(res, req)
	case http.MethodDelete:
		s.deleteSuiteCollectionHandler(res, req)
	default:
		log.Panicf("method '%s' not implemented\n", req.Method)
	}
}

func (s *srv) getSuiteCollectionHandler(res http.ResponseWriter, req *http.Request) {
	formValSince := req.FormValue("since")
	var sinceTime int64
	if formValSince != "" {
		var err error
		sinceTime, err = strconv.ParseInt(formValSince, 10, 64)
		if err != nil {
			httpError(res, errBadQuery, http.StatusBadRequest)
			return
		}
	}

	suiteRuns, err := s.db.WithContext(req.Context()).AllSuiteRuns(sinceTime)
	if err != nil {
		log.Printf("get all suite runs: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
		return
	}
	httpJson(res, suiteRuns, http.StatusOK)
}

func (s *srv) postSuiteCollectionHandler(res http.ResponseWriter, req *http.Request) {
	var suiteRun database.NewSuiteRun
	if err := json.NewDecoder(req.Body).Decode(&suiteRun); err != nil {
		httpError(res, errBadJson, http.StatusBadRequest)
		return
	}
	id, err := s.db.WithContext(req.Context()).NewSuiteRun(suiteRun)
	if errors.Is(err, database.ErrInvalidModel) {
		httpError(res, errBadJson, http.StatusBadRequest)
		return
	} else if err != nil {
		log.Printf("create suite run: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
		return
	}

	loc, err := s.router.Get("suite").URL("suite_id", id)
	if err != nil {
		log.Printf("build suite url: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
		return
	}

	res.Header().Set("Location", loc.String())
	httpJson(res, bson.M{"id": id}, http.StatusCreated)
}

func (s *srv) deleteSuiteCollectionHandler(res http.ResponseWriter, req *http.Request) {
	err := s.db.WithContext(req.Context()).DeleteAllSuiteRuns()
	if err != nil {
		log.Printf("delete all suite runs: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusNoContent)
}
