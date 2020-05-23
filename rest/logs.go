package rest

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *srv) getLogHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		logMsg, err := s.repos.Logs().Find(r.Context(), id)
		if err != nil {
			return fmt.Errorf("get log message: %v", err)
		}
		return writeJson(w, http.StatusOK, logMsg)
	})
}

func (s *srv) getLogCollectionHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		caseId := mux.Vars(r)["case_id"]
		logMsgs, err := s.repos.Logs().FindAllByCase(r.Context(), caseId)
		if err != nil {
			return fmt.Errorf("get all log messages for case: %v", err)
		}
		return writeJson(w, http.StatusOK, logMsgs)
	})
}
