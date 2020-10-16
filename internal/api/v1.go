package api

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/suiteserve/suiteserve/internal/repo"
	"github.com/suiteserve/suiteserve/sse"
	"io"
	"log"
	"net/http"
	"time"
)

type Repo interface {
	InsertAttachment(ctx context.Context, a repo.Attachment) (id repo.Id, err error)
	Attachment(ctx context.Context, id repo.Id) (interface{}, error)
	AllAttachments(ctx context.Context) (interface{}, error)
	SuiteAttachments(ctx context.Context, suiteId repo.Id) (interface{}, error)
	CaseAttachments(ctx context.Context, caseId repo.Id) (interface{}, error)

	InsertSuite(ctx context.Context, s repo.Suite) (id repo.Id, err error)
	Suite(ctx context.Context, id repo.Id) (interface{}, error)
	SuitePage(ctx context.Context) (interface{}, error)
	SuitePageAfter(ctx context.Context, id repo.Id) (interface{}, error)
	WatchSuites(ctx context.Context) *repo.Watcher
	DeleteSuite(ctx context.Context, id repo.Id, at int64) error
	FinishSuite(ctx context.Context, id repo.Id, result repo.SuiteResult, at int64) error
	DisconnectSuite(ctx context.Context, id repo.Id, at int64) error

	InsertCase(ctx context.Context, c repo.Case) (id repo.Id, err error)
	Case(ctx context.Context, id repo.Id) (interface{}, error)
	SuiteCases(ctx context.Context, suiteId repo.Id) (interface{}, error)

	InsertLogLine(ctx context.Context, ll repo.LogLine) (id repo.Id, err error)
	LogLine(ctx context.Context, id repo.Id) (interface{}, error)
	CaseLogLines(ctx context.Context, llId repo.Id) (interface{}, error)
}

type v1 struct {
	repo Repo
}

func NewV1Handler(r Repo) http.Handler {
	return v1{r}.newRouter()
}

func (v v1) newRouter() http.Handler {
	r := mux.NewRouter()
	r.NotFoundHandler = notFound()
	r.MethodNotAllowedHandler = methodNotAllowed()

	// attachments
	r.Handle("/attachments", findByIdHandler(v.repo.SuiteAttachments)).
		Queries("suite", "{id}").
		Methods(http.MethodGet, http.MethodHead)
	r.Handle("/attachments", findByIdHandler(v.repo.CaseAttachments)).
		Queries("case", "{id}").
		Methods(http.MethodGet, http.MethodHead)
	r.Handle("/attachments", findAllHandler(v.repo.AllAttachments)).
		Methods(http.MethodGet, http.MethodHead)
	r.Handle("/attachments/{id}", findByIdHandler(v.repo.Attachment)).
		Methods(http.MethodGet, http.MethodHead)

	// suites
	r.Handle("/suites", v.insertSuiteHandler()).
		Methods(http.MethodPost)
	r.Handle("/suites", findByIdHandler(v.repo.SuitePageAfter)).
		Queries("after", "{id}").
		Methods(http.MethodGet, http.MethodHead)
	r.Handle("/suites", sse.NewMiddleware(v.watchSuitesHandler())).
		Queries("watch", "true").
		Methods(http.MethodGet, http.MethodHead)
	r.Handle("/suites", findAllHandler(v.repo.SuitePage)).
		Methods(http.MethodGet, http.MethodHead)
	r.Handle("/suites/{id}", findByIdHandler(v.repo.Suite)).
		Methods(http.MethodGet, http.MethodHead)
	r.Handle("/suites/{id}/cases", findByIdHandler(v.repo.SuiteCases)).
		Methods(http.MethodGet, http.MethodHead)

	// cases
	r.Handle("/cases/{id}", findByIdHandler(v.repo.Case)).
		Methods(http.MethodGet, http.MethodHead)
	r.Handle("/cases/{id}/logs", findByIdHandler(v.repo.CaseLogLines)).
		Methods(http.MethodGet, http.MethodHead)

	// logs
	r.Handle("/logs/{id}", findByIdHandler(v.repo.LogLine)).
		Methods(http.MethodGet, http.MethodHead)

	return r
}

func (v *v1) insertSuiteHandler() errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var s repo.Suite
		if err := readJson(r, &s); err != nil {
			return err
		}
		id, err := v.repo.InsertSuite(r.Context(), s)
		if err != nil {
			return err
		}
		return writeJson(w, r, id)
	}
}

func (v *v1) insertCaseHandler() errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var c repo.Case
		if err := readJson(r, &c); err != nil {
			return err
		}
		id, err := v.repo.InsertCase(r.Context(), c)
		if err != nil {
			return err
		}
		return writeJson(w, r, id)
	}
}

func (v *v1) insertLogLineHandler() errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var ll repo.LogLine
		if err := readJson(r, &ll); err != nil {
			return err
		}
		id, err := v.repo.InsertLogLine(r.Context(), ll)
		if err != nil {
			return err
		}
		return writeJson(w, r, id)
	}
}

func (v *v1) watchSuitesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		watcher := v.repo.WatchSuites(r.Context())
		var err error
		for ok := true; ok; {
			if ok, err = v.watchSuitesTick(w, watcher); err != nil {
				log.Printf("<%s> %v", r.RemoteAddr, err)
				return
			}
		}
	}
}

func (v *v1) watchSuitesTick(w io.Writer, watcher *repo.Watcher) (bool, error) {
	timer := time.NewTimer(15 * time.Second)
	defer timer.Stop()
	select {
	case c, ok := <-watcher.Ch():
		if !ok {
			return false, watcher.Err()
		}
		_, err := sse.Send(w, sse.WithEventType(c.Coll),
			sse.WithData(string(c.Msg)))
		if err != nil {
			return false, err
		}
	case <-timer.C:
		if _, err := sse.Send(w, sse.WithComment("keep-alive")); err != nil {
			return false, err
		}
	}
	return true, nil
}

func findHandler(fn func(r *http.Request) (interface{}, error)) errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		v, err := fn(r)
		if err != nil {
			return err
		}
		return writeJson(w, r, v)
	}
}

func findAllHandler(fn func(ctx context.Context) (interface{}, error)) errHandlerFunc {
	return findHandler(func(r *http.Request) (interface{}, error) {
		return fn(r.Context())
	})
}

func findByIdHandler(fn func(ctx context.Context, id repo.Id) (interface{}, error)) errHandlerFunc {
	return findHandler(func(r *http.Request) (interface{}, error) {
		id, err := parseIdVar(r, "id")
		if err != nil {
			return nil, err
		}
		return fn(r.Context(), id)
	})
}
