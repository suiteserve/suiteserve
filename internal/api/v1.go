package api

import (
	"github.com/suiteserve/suiteserve/internal/repo"
	"github.com/suiteserve/suiteserve/middleware"
	"net/http"
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
		middleware.Param("/attachments/",
			v1.attachment()))
	mux.Handle("/attachments",
		v1.attachmentCollection())
	mux.Handle("/", middleware.NotFound())
	return middleware.Methods(http.MethodGet, http.MethodHead)(&mux)
}

func (v *v1) attachment() http.HandlerFunc {
	return errHandler(func(w http.ResponseWriter, r *http.Request) error {
		id := middleware.GetParam(r)
		a, err := v.repo.Attachment(id)
		if err != nil {
			return err
		}
		return writeJson(w, r, &a)
	})
}

func (v *v1) attachmentCollection() http.HandlerFunc {
	return errHandler(func(w http.ResponseWriter, r *http.Request) error {
		var a interface{}
		var err error
		if suiteId := r.FormValue("suite"); suiteId != "" {
			a, err = v.repo.SuiteAttachments(suiteId)
		} else if caseId := r.FormValue("case"); caseId != "" {
			a, err = v.repo.CaseAttachments(caseId)
		} else {
			return httpError{code: http.StatusBadRequest}
		}
		if err != nil {
			return err
		}
		return writeJson(w, r, a)
	})
}
