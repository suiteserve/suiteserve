package api

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/suiteserve/suiteserve/internal/repo"
	"github.com/suiteserve/suiteserve/sse"
	"net/http"
	"time"
)

type Repo interface {
	InsertAttachment(ctx context.Context, a repo.Attachment) (id repo.Id, err error)
	Attachment(ctx context.Context, id repo.Id) (repo.Attachment, error)
	AllAttachments(ctx context.Context) ([]repo.Attachment, error)
	SuiteAttachments(ctx context.Context, suiteId repo.Id) ([]repo.Attachment, error)
	CaseAttachments(ctx context.Context, caseId repo.Id) ([]repo.Attachment, error)

	InsertSuite(ctx context.Context, s repo.Suite) (id repo.Id, err error)
	Suite(ctx context.Context, id repo.Id) (repo.Suite, error)
	SuitePage(ctx context.Context) (repo.SuitePage, error)
	SuitePageAfter(ctx context.Context, cursor repo.SuitePageCursor) (repo.SuitePage, error)
	FinishSuite(ctx context.Context, id repo.Id, result repo.SuiteResult, at repo.MsTime) error
	DisconnectSuite(ctx context.Context, id repo.Id, at repo.MsTime) error

	InsertCase(ctx context.Context, c repo.Case) (id repo.Id, err error)
	Case(ctx context.Context, id repo.Id) (repo.Case, error)
	SuiteCases(ctx context.Context, suiteId repo.Id) ([]repo.Case, error)
	FinishCase(ctx context.Context, id repo.Id, result repo.CaseResult, at repo.MsTime) error

	InsertLogLine(ctx context.Context, ll repo.LogLine) (id repo.Id, err error)
	LogLine(ctx context.Context, id repo.Id) (repo.LogLine, error)
	CaseLogLines(ctx context.Context, llId repo.Id) ([]repo.LogLine, error)

	Watch(ctx context.Context) (<-chan repo.Change, <-chan error)
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
	r.Handle("/attachments", findByIdHandler(func(ctx context.Context, id repo.Id) (interface{}, error) {
		return v.repo.SuiteAttachments(ctx, id)
	})).
		Queries("suite", "{id}").
		Methods(http.MethodGet, http.MethodHead)
	r.Handle("/attachments", findByIdHandler(func(ctx context.Context, id repo.Id) (interface{}, error) {
		return v.repo.CaseAttachments(ctx, id)
	})).
		Queries("case", "{id}").
		Methods(http.MethodGet, http.MethodHead)
	r.Handle("/attachments", findAllHandler(func(ctx context.Context) (interface{}, error) {
		return v.repo.AllAttachments(ctx)
	})).
		Methods(http.MethodGet, http.MethodHead)
	r.Handle("/attachments/{id}", findByIdHandler(func(ctx context.Context, id repo.Id) (interface{}, error) {
		return v.repo.Attachment(ctx, id)
	})).
		Methods(http.MethodGet, http.MethodHead)

	// suites
	r.Handle("/suites/{id}/cases", findByIdHandler(func(ctx context.Context, id repo.Id) (interface{}, error) {
		return v.repo.SuiteCases(ctx, id)
	})).
		Methods(http.MethodGet, http.MethodHead)
	r.Handle("/suites/{id}", v.finishSuiteHandler()).
		Queries("finish", "true").
		Methods(http.MethodPatch)
	r.Handle("/suites/{id}", findByIdHandler(func(ctx context.Context, id repo.Id) (interface{}, error) {
		return v.repo.Suite(ctx, id)
	})).
		Methods(http.MethodGet, http.MethodHead)
	r.Handle("/suites", findHandler(func(r *http.Request) (interface{}, error) {
		cursorStr := getVar(r, "cursor")
		cursor, err := repo.NewSuitePageCursor(cursorStr)
		if err != nil {
			return nil, err
		}
		return v.repo.SuitePageAfter(r.Context(), cursor)
	})).
		Queries("from", "{cursor}").
		Methods(http.MethodGet, http.MethodHead)
	r.Handle("/suites", sse.NewMiddleware(v.watchHandler())).
		Queries("watch", "true").
		Methods(http.MethodGet, http.MethodHead)
	r.Handle("/suites", findAllHandler(func(ctx context.Context) (interface{}, error) {
		return v.repo.SuitePage(ctx)
	})).
		Methods(http.MethodGet, http.MethodHead)
	r.Handle("/suites", v.insertSuiteHandler()).
		Methods(http.MethodPost)

	// cases
	r.Handle("/cases/{id}/logs", findByIdHandler(func(ctx context.Context, id repo.Id) (interface{}, error) {
		return v.repo.CaseLogLines(ctx, id)
	})).
		Methods(http.MethodGet, http.MethodHead)
	r.Handle("/cases/{id}", v.finishCaseHandler()).
		Queries("finish", "true").
		Methods(http.MethodPatch)
	r.Handle("/cases/{id}", findByIdHandler(func(ctx context.Context, id repo.Id) (interface{}, error) {
		return v.repo.Case(ctx, id)
	})).
		Methods(http.MethodGet, http.MethodHead)
	r.Handle("/cases", v.insertCaseHandler()).
		Methods(http.MethodPost)

	// logs
	r.Handle("/logs", v.insertLogLineHandler()).
		Methods(http.MethodPost)
	r.Handle("/logs/{id}", findByIdHandler(func(ctx context.Context, id repo.Id) (interface{}, error) {
		return v.repo.LogLine(ctx, id)
	})).
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

func (v *v1) finishSuiteHandler() errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		id, err := getIdVar(r)
		if err != nil {
			return err
		}
		var in struct {
			Result repo.SuiteResult `json:"result"`
			At     repo.MsTime      `json:"at"`
		}
		if err := readJson(r, &in); err != nil {
			return err
		}
		return v.repo.FinishSuite(r.Context(), id, in.Result, in.At)
	}
}

func (v *v1) finishCaseHandler() errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		id, err := getIdVar(r)
		if err != nil {
			return err
		}
		var in struct {
			Result repo.CaseResult `json:"result"`
			At     repo.MsTime     `json:"at"`
		}
		if err := readJson(r, &in); err != nil {
			return err
		}
		return v.repo.FinishCase(r.Context(), id, in.Result, in.At)
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

func (v *v1) watchHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		changeCh, errCh := v.repo.Watch(r.Context())
		for evts := range changesToSSE(changeCh) {
			if _, err := sse.Send(w, evts...); err != nil {
				printLog(r, err)
				return
			}
		}
		if err := <-errCh; err != nil {
			printLog(r, err)
			return
		}
	}
}

func changesToSSE(ch <-chan repo.Change) <-chan []sse.Event {
	out := make(chan []sse.Event)
	go func() {
		defer close(out)
		for {
			timer := time.NewTimer(15 * time.Second)
			select {
			case c, ok := <-ch:
				if !timer.Stop() {
					<-timer.C
				}
				if !ok {
					return
				}
				out <- []sse.Event{
					sse.WithEventType(string(c.Coll)),
					sse.WithData(string(c.Msg)),
				}
			case <-timer.C:
				out <- []sse.Event{
					sse.WithComment("keep-alive"),
				}
			}
		}
	}()
	return out
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
		id, err := getIdVar(r)
		if err != nil {
			return nil, err
		}
		return fn(r.Context(), id)
	})
}
