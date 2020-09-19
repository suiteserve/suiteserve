package api

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/suiteserve/suiteserve/internal/repo"
	"net/http"
)

type Repo interface {
	InsertAttachment(ctx context.Context, a repo.Attachment) (id repo.Id, err error)
	Attachment(ctx context.Context, id repo.Id) (*repo.Attachment, error)
	Attachments(ctx context.Context) ([]repo.Attachment, error)
	SuiteAttachments(ctx context.Context, suiteId repo.Id) ([]repo.Attachment, error)
	CaseAttachments(ctx context.Context, caseId repo.Id) ([]repo.Attachment, error)

	InsertSuite(ctx context.Context, s repo.Suite) (id repo.Id, err error)
	Suite(ctx context.Context, id repo.Id) (*repo.Suite, error)
	// SuitePage(ctx context.Context) (*repo.SuitePage, error)
	// SuitePageTo(ctx context.Context, id repo.Id) (*repo.SuitePage, error)
	// SuitePageFrom(ctx context.Context, id repo.Id) (*repo.SuitePage, error)
	// SuitePageAround(ctx context.Context, id repo.Id) (*repo.SuitePage, error)
	WatchSuites(ctx context.Context) *repo.Watcher
	DeleteSuite(ctx context.Context, id repo.Id, at int64) error
	FinishSuite(ctx context.Context, id repo.Id, result repo.SuiteResult, at int64) error
	DisconnectSuite(ctx context.Context, id repo.Id, at int64) error

	InsertCase(ctx context.Context, c repo.Case) (id repo.Id, err error)
	Case(ctx context.Context, id repo.Id) (*repo.Case, error)

	InsertLogLine(ctx context.Context, ll repo.LogLine) (id repo.Id, err error)
	LogLine(ctx context.Context, id repo.Id) (*repo.LogLine, error)
}

type v1 struct {
	repo Repo
}

func NewV1Handler(repo Repo) http.Handler {
	v1 := v1{repo}
	r := mux.NewRouter()
	r.NotFoundHandler = notFound()
	r.MethodNotAllowedHandler = methodNotAllowed()
	r.Handle("/attachments", v1.suiteAttachmentsHandler()).
		Queries("suite", "{suiteId}")
	r.Handle("/attachments", v1.caseAttachmentsHandler()).
		Queries("case", "{caseId}")
	r.Handle("/attachments", v1.allAttachmentsHandler())
	r.Handle("/attachments/{id}", v1.attachmentHandler())
	r.Handle("/suites", v1.watchSuitesHandler()).
		Queries("watch", "{watch}", "from", "{fromId}")
	r.Handle("/suites/{id}", v1.suiteHandler())
	// r.Handle("/suites/{id}/cases", v1.suiteCasesHandler())
	r.Handle("/cases/{id}", v1.caseHandler())
	// r.Handle("/cases/{id}/logs", v1.caseLogsHandler())
	r.Handle("/logs/{id}", v1.logLineHandler())
	return methodsMw(http.MethodGet, http.MethodHead)(r)
}

func (v *v1) suiteAttachmentsHandler() errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		suiteId, err := hexVarToId(r, "suiteId")
		if err != nil {
			return err
		}
		a, err := v.repo.SuiteAttachments(r.Context(), suiteId)
		if err != nil {
			return err
		}
		return writeJson(w, r, a)
	}
}

func (v *v1) caseAttachmentsHandler() errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		caseId, err := hexVarToId(r, "caseId")
		if err != nil {
			return err
		}
		a, err := v.repo.CaseAttachments(r.Context(), caseId)
		if err != nil {
			return err
		}
		return writeJson(w, r, a)
	}
}

func (v *v1) allAttachmentsHandler() errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		a, err := v.repo.Attachments(r.Context())
		if err != nil {
			return err
		}
		return writeJson(w, r, a)
	}
}

func (v *v1) attachmentHandler() errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		id, err := hexVarToId(r, "id")
		if err != nil {
			return err
		}
		a, err := v.repo.Attachment(r.Context(), id)
		if err != nil {
			return err
		}
		return writeJson(w, r, a)
	}
}

func (v *v1) watchSuitesHandler() errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// fromId, err := hexVarToId(r, "from")
		// if err != nil {
		// 	return err
		// }
		// v.repo.Suite
		// s, err := v.repo.Suite(r.Context(), id)
		// if err != nil {
		// 	return err
		// }
		return writeJson(w, r, nil)
	}
}

func (v *v1) suiteHandler() errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		id, err := hexVarToId(r, "id")
		if err != nil {
			return err
		}
		s, err := v.repo.Suite(r.Context(), id)
		if err != nil {
			return err
		}
		return writeJson(w, r, s)
	}
}

// func suiteWatchWriter(ctx context.Context, w io.Writer,
// 	watcher *repo.SuiteWatcher) (bool, error) {
// 	timer := time.NewTimer(15 * time.Second)
// 	defer timer.Stop()
// 	select {
// 	case changes := <-watcher.Changes():
// 		for _, c := range changes {
// 			b, err := json.Marshal(c)
// 			if err != nil {
// 				panic(err)
// 			}
// 			_, err = sse.Send(w, sse.WithEventType(c.Type()),
// 				sse.WithData(string(b)))
// 			if err != nil {
// 				return false, err
// 			}
// 		}
// 	case <-timer.C:
// 		if _, err := sse.Send(w, sse.WithComment("keep-alive")); err != nil {
// 			return false, err
// 		}
// 	case <-ctx.Done():
// 		return false, nil
// 	}
// 	return true, nil
// }

func (v *v1) caseHandler() errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		id, err := hexVarToId(r, "id")
		if err != nil {
			return err
		}
		c, err := v.repo.Case(r.Context(), id)
		if err != nil {
			return err
		}
		return writeJson(w, r, c)
	}
}

func (v *v1) logLineHandler() errHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		id, err := hexVarToId(r, "id")
		if err != nil {
			return err
		}
		ll, err := v.repo.LogLine(r.Context(), id)
		if err != nil {
			return err
		}
		return writeJson(w, r, ll)
	}
}
