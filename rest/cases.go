package rest

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *srv) getCaseHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		c, err := s.repos.Cases().Find(r.Context(), id)
		if err != nil {
			return fmt.Errorf("get case run: %v", err)
		}
		return writeJson(w, http.StatusOK, c)
	})
}

func (s *srv) getCaseCollectionHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		suiteId := mux.Vars(r)["suite_id"]
		num, err := parseInt64(r.FormValue("num"))
		if err != nil {
			return errBadQuery(err)
		}

		cases, err := s.repos.Cases().FindAllBySuite(r.Context(), suiteId, num)
		if err != nil {
			return fmt.Errorf("get all cases for suite: %v", err)
		}
		return writeJson(w, http.StatusOK, cases)
	})
}
