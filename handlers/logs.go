package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tmazeika/testpass/database"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
)

func (s *srv) logHandler(res http.ResponseWriter, req *http.Request) error {
	return oneArgHandlerMap{
		http.MethodGet: s.getLogHandler,
	}.handle(res, req, "log_id")
}

func (s *srv) getLogHandler(res http.ResponseWriter, req *http.Request, id string) error {
	logMsg, err := s.db.WithContext(req.Context()).LogMessage(id)
	if errors.Is(err, database.ErrNotFound) {
		return errNotFound
	} else if err != nil {
		return fmt.Errorf("get log message: %v", err)
	}

	writeJson(res, logMsg, http.StatusOK)
	return nil
}

func (s *srv) logCollectionHandler(res http.ResponseWriter, req *http.Request) error {
	return oneArgHandlerMap{
		http.MethodGet:  s.getLogCollectionHandler,
		http.MethodPost: s.postLogCollectionHandler,
	}.handle(res, req, "case_id")
}

func (s *srv) getLogCollectionHandler(res http.ResponseWriter, req *http.Request, caseId string) error {
	logMsgs, err := s.db.WithContext(req.Context()).AllLogMessages(caseId)
	if err != nil {
		return fmt.Errorf("get all log messages for case: %v", err)
	}

	writeJson(res, logMsgs, http.StatusOK)
	return nil
}

func (s *srv) postLogCollectionHandler(res http.ResponseWriter, req *http.Request, caseId string) error {
	var logMsg database.NewLogMessage
	if err := json.NewDecoder(req.Body).Decode(&logMsg); err != nil {
		return errBadJson
	}

	id, err := s.db.WithContext(req.Context()).NewLogMessage(caseId, logMsg)
	if errors.Is(err, database.ErrInvalidModel) {
		return errBadJson
	} else if err != nil {
		return fmt.Errorf("new log message: %v", err)
	}
	//TODO go s.publishLogEvent(eventTypeCreateLog, id)

	loc, err := s.router.Get("log").URL("log_id", id)
	if err != nil {
		return fmt.Errorf("build log url: %v", err)
	}

	res.Header().Set("Location", loc.String())
	writeJson(res, bson.M{"id": id}, http.StatusCreated)
	return nil
}

func (s *srv) publishLogEvent(eType eventType, id string) {
	logMsg, err := s.db.WithContext(context.Background()).LogMessage(id)
	if err != nil {
		log.Printf("get log message: %v", err)
	} else {
		s.eventBus.publish(newEvent(eType, logMsg))
	}
}
