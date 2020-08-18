package api

import (
	"context"
	"encoding/json"
	"github.com/suiteserve/suiteserve/internal/repo"
	"github.com/suiteserve/suiteserve/sse"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Repo interface {
	InsertAttachment(repo.Attachment) (id string, err error)
	Attachment(id string) (repo.Attachment, error)
	SuiteAttachments(suiteId string) ([]repo.Attachment, error)
	CaseAttachments(caseId string) ([]repo.Attachment, error)

	InsertSuite(repo.Suite) (id string, err error)
	Suite(id string) (repo.Suite, error)
	WatchSuites(id string, padLt, padGt int) (*repo.SuiteWatcher, error)

	InsertCase(repo.Case) (id string, err error)
	Case(id string) (repo.Case, error)

	InsertLogLine(repo.LogLine) (id string, err error)
	LogLine(id string) (repo.LogLine, error)
}

type v1 struct {
	repo Repo
}

func NewV1Handler(repo Repo) http.Handler {
	v1 := v1{repo}
	var mux http.ServeMux
	mux.Handle("/attachments/",
		pathParamMw("/attachments/", v1.attachmentHandler()))
	mux.Handle("/attachments",
		v1.attachmentCollHandler())
	mux.Handle("/suites/",
		pathParamMw("/suites/", v1.suiteHandler()))
	mux.Handle("/suites",
		sse.NewMiddleware(v1.suiteCollHandler()))
	mux.Handle("/cases/",
		pathParamMw("/cases/", v1.caseHandler()))
	mux.Handle("/logs/",
		pathParamMw("/logs/", v1.logLineHandler()))
	mux.Handle("/", notFound())
	return methodsMw(http.MethodGet, http.MethodHead)(&mux)
}

func (v *v1) attachmentHandler() errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		id := pathParam(r)
		a, err := v.repo.Attachment(id)
		if err != nil {
			return err
		}
		return writeJson(w, r, &a)
	}
}

func (v *v1) attachmentCollHandler() errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var a interface{}
		var err error
		if suiteId := r.FormValue("suite"); suiteId != "" {
			a, err = v.repo.SuiteAttachments(suiteId)
		} else if caseId := r.FormValue("case"); caseId != "" {
			a, err = v.repo.CaseAttachments(caseId)
		} else {
			return errHttp{code: http.StatusBadRequest}
		}
		if err != nil {
			return err
		}
		return writeJson(w, r, a)
	}
}

func (v *v1) suiteHandler() errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		id := pathParam(r)
		s, err := v.repo.Suite(id)
		if err != nil {
			return err
		}
		return writeJson(w, r, &s)
	}
}

func (v *v1) suiteCollHandler() errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		padGtStr := r.FormValue("pad_gt")
		id := r.FormValue("id")
		padLtStr := r.FormValue("pad_lt")
		padGt, err := strconv.ParseInt(padGtStr, 10, 32)
		if err != nil {
			return errHttp{code: http.StatusBadRequest, cause: err}
		}
		padLt, err := strconv.ParseInt(padLtStr, 10, 32)
		if err != nil {
			return errHttp{code: http.StatusBadRequest, cause: err}
		}

		watcher, err := v.repo.WatchSuites(id, int(padLt), int(padGt))
		if err != nil {
			return err
		}
		defer watcher.Close()

		for {
			if ok, err := suiteWatchWriter(r.Context(), w, watcher); !ok {
				return err
			}
		}
	}
}

func suiteWatchWriter(ctx context.Context, w io.Writer,
	watcher *repo.SuiteWatcher) (bool, error) {
	timer := time.NewTimer(15 * time.Second)
	defer timer.Stop()
	select {
	case changes := <-watcher.Changes():
		for _, c := range changes {
			b, err := json.Marshal(c)
			if err != nil {
				panic(err)
			}
			_, err = sse.Send(w, sse.WithEventType(c.Type()),
				sse.WithData(string(b)))
			if err != nil {
				return false, err
			}
		}
	case <-timer.C:
		if _, err := sse.Send(w, sse.WithComment("keep-alive")); err != nil {
			return false, err
		}
	case <-ctx.Done():
		return false, nil
	}
	return true, nil
}

func (v *v1) caseHandler() errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		id := pathParam(r)
		c, err := v.repo.Case(id)
		if err != nil {
			return err
		}
		return writeJson(w, r, &c)
	}
}

func (v *v1) logLineHandler() errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		id := pathParam(r)
		ll, err := v.repo.LogLine(id)
		if err != nil {
			return err
		}
		return writeJson(w, r, &ll)
	}
}
