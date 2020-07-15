package rest

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/suiteserve/suiteserve/repo"
	"net/http"
	"strconv"
)

type logFinder interface {
	LogPage(ctx context.Context, suiteId string, fromId string, limit int) (*repo.LogPage, error)
}

func newGetLogCollectionHandler(finder logFinder) http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		suiteId := mux.Vars(r)["suite_id"]
		fromId := r.FormValue("from_id")
		limit, err := strconv.ParseInt(r.FormValue("limit"), 10, 32)
		if err != nil || limit < 1 {
			limit = 10
		}

		logs, err := finder.LogPage(r.Context(), suiteId, fromId, int(limit))
		if errors.Is(err, repo.ErrNotFound) {
			return errNotFound(err)
		} else if err != nil {
			return fmt.Errorf("get logs page: %v", err)
		}
		return writeJson(w, http.StatusOK, logs)
	})
}
