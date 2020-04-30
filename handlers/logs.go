package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/tmazeika/testpass/database"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
)

func (s *srv) logHandler(res http.ResponseWriter, req *http.Request) {
	logId, ok := mux.Vars(req)["log_id"]
	if !ok {
		log.Panicln("req param 'log_id' not found")
	}

	switch req.Method {
	case http.MethodGet:
		s.getLogHandler(res, req, logId)
	default:
		log.Panicf("method '%s' not implemented\n", req.Method)
	}
}

func (s *srv) getLogHandler(res http.ResponseWriter, req *http.Request, logId string) {
	logMsg, err := s.db.WithContext(req.Context()).LogMessage(logId)
	if errors.Is(err, database.ErrNotFound) {
		httpError(res, errNotFound, http.StatusNotFound)
	} else if err != nil {
		log.Printf("get log message: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
	} else {
		httpJson(res, logMsg, http.StatusOK)
	}
}

func (s *srv) logCollectionHandler(res http.ResponseWriter, req *http.Request) {
	caseId, ok := mux.Vars(req)["case_id"]
	if !ok {
		log.Panicln("req param 'case_id' not found")
	}

	switch req.Method {
	case http.MethodGet:
		s.getLogCollectionHandler(res, req, caseId)
	case http.MethodPost:
		s.postLogCollectionHandler(res, req, caseId)
	default:
		log.Panicf("method '%s' not implemented\n", req.Method)
	}
}

func (s *srv) getLogCollectionHandler(res http.ResponseWriter, req *http.Request, caseId string) {
	logMsgs, err := s.db.WithContext(req.Context()).AllLogMessages(caseId)
	if err != nil {
		log.Printf("get all log messages for case run: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
	} else {
		httpJson(res, logMsgs, http.StatusOK)
	}
}

func (s *srv) postLogCollectionHandler(res http.ResponseWriter, req *http.Request, caseId string) {
	var logMsg database.NewLogMessage
	if err := json.NewDecoder(req.Body).Decode(&logMsg); err != nil {
		httpError(res, errBadJson, http.StatusBadRequest)
		return
	}
	id, err := s.db.WithContext(req.Context()).NewLogMessage(caseId, logMsg)
	if errors.Is(err, database.ErrInvalidModel) {
		httpError(res, errBadJson, http.StatusBadRequest)
		return
	} else if err != nil {
		log.Printf("create log message: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
		return
	}
	go s.publishLogEvent(eventTypeCreateLog, id)

	loc, err := s.router.Get("log").URL("log_id", id)
	if err != nil {
		log.Printf("build log url: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
		return
	}

	res.Header().Set("Location", loc.String())
	httpJson(res, bson.M{"id": id}, http.StatusCreated)
}

func (s *srv) publishLogEvent(eType eventType, id string) {
	logMsg, err := s.db.WithContext(context.Background()).LogMessage(id)
	if err != nil {
		log.Printf("get log message: %v\n", err)
	} else {
		s.eventBus.publish(newEvent(eType, logMsg))
	}
}
