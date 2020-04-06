package handlers

import (
	"github.com/gorilla/mux"
	"github.com/tmazeika/testpass/database"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"log"
	"net/http"
)

func (s *srv) suiteHandler(res http.ResponseWriter, req *http.Request) {
	id, ok := mux.Vars(req)["suiteId"]
	if !ok {
		panic("request parameter 'suiteId' not found")
	}

	switch req.Method {
	case http.MethodGet:
		suiteRun, err := s.db.SuiteRun(id)
		if err == database.ErrNotFound {
			httpError(res, errSuiteRunNotFound, http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("failed to get suite run: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}
		httpJson(res, suiteRun, http.StatusOK)
	case http.MethodDelete:
		//	if err := s.db.DeleteAttachment(id); err == database.ErrNotFound {
		//		httpError(res, errAttachmentNotFound, http.StatusNotFound)
		//		return
		//	} else if err != nil {
		//		log.Printf("failed to delete attachment: %v\n", err)
		//		httpError(res, errUnknown, http.StatusInternalServerError)
		//		return
		//	}
		//	res.WriteHeader(http.StatusNoContent)
	}
}

func (s *srv) suitesHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		//attachments, err := s.db.AllAttachments()
		//if err != nil {
		//	log.Printf("failed to get attachments: %v\n", err)
		//	httpError(res, errUnknown, http.StatusInternalServerError)
		//	return
		//}
		//httpJson(res, attachments, http.StatusOK)
	case http.MethodPost:
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Printf("failed to read HTTP body: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}

		id, err := s.db.NewSuiteRun(b)
		if err == database.ErrBadJson {
			httpError(res, errBadJson, http.StatusBadRequest)
			return
		} else if err != nil {
			log.Printf("failed to create new suite run: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}

		loc, err := s.router.Get("suite").URL("suiteId", id)
		if err != nil {
			log.Printf("failed to build URL to suite: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}

		res.Header().Set("Location", loc.String())
		httpJson(res, bson.M{"id": id}, http.StatusCreated)
	case http.MethodDelete:
		//if err := s.db.DeleteAllAttachments(); err != nil {
		//	log.Printf("failed to delete attachments: %v\n", err)
		//	httpError(res, errUnknown, http.StatusInternalServerError)
		//	return
		//}
		//res.WriteHeader(http.StatusNoContent)
	}
}
