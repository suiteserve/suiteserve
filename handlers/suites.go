package handlers

import (
	"github.com/gorilla/mux"
	"github.com/tmazeika/testpass/database"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

func (s *srv) suiteHandler(res http.ResponseWriter, req *http.Request) {
	id, ok := mux.Vars(req)["suite_id"]
	if !ok {
		panic("request parameter 'suite_id' not found")
	}

	switch req.Method {
	case http.MethodGet:
		suiteRun, err := s.db.SuiteRun(id)
		if err == database.ErrNotFound {
			httpError(res, errNotFound, http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("get suite run: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}
		httpJson(res, suiteRun, http.StatusOK)
	case http.MethodDelete:
		if err := s.db.DeleteSuiteRun(id); err == database.ErrNotFound {
			httpError(res, errNotFound, http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("delete suite run: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusNoContent)
	}
}

func (s *srv) suitesHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		sinceStr := req.FormValue("since")
		var sinceTime time.Time
		if sinceStr != "" {
			sinceInt, err := strconv.ParseInt(sinceStr, 10, 64)
			if err != nil {
				httpError(res, errBadQuery, http.StatusBadRequest)
				return
			}
			sinceTime = time.Unix(sinceInt, 0)
		}

		suiteRuns, err := s.db.AllSuiteRuns(sinceTime)
		if err != nil {
			log.Printf("get suite runs: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}
		httpJson(res, suiteRuns, http.StatusOK)
	case http.MethodPost:
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Printf("read http body: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}

		id, err := s.db.NewSuiteRun(b)
		if err == database.ErrBadJson {
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
	case http.MethodDelete:
		if err := s.db.DeleteAllSuiteRuns(); err != nil {
			log.Printf("delete suite runs: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusNoContent)
	}
}
