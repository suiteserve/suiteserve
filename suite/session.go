package suite

import (
	"encoding/json"
	"log"
	"net"
)

type session struct {
	*Server
	enc *json.Encoder
	id  string
}

func (s *Server) newSession(conn net.Conn) *session {
	return &session{
		Server: s,
		enc:    json.NewEncoder(conn),
	}
}

func (h *session) detach() error {
	if h.id != "" {
		h.detachedSuites.detach(h.id)
	}
	// TODO: update suite status
	return nil
}

func (h *session) helloHandler(r *request) (handler, error) {
	cmd, ok := r.obj["cmd"].(string)
	if !ok {
		return nil, errBadCmd(r.seq, r.obj["cmd"], "not a string")
	}
	switch cmd {
	case "start":
		return h.startHandler(r)
	case "reattach":
		return h.reattachHandler(r)
	default:
		return nil, errBadCmd(r.seq, cmd, "not one of [start, reattach]")
	}
}

func (h *session) startHandler(r *request) (handler, error) {
	suite, ok := r.obj["suite"].(map[string]interface{})
	if !ok {
		return nil, errBadSuite(r.seq, r.obj["suite"], "not an object")
	}

	// TODO
	log.Println("DEBUG", suite)
	return nil, nil
}

func (h *session) reattachHandler(r *request) (handler, error) {
	id, ok := r.obj["id"].(string)
	if !ok {
		return nil, errSuiteNotFound(r.seq, r.obj["id"])
	}
	if !h.detachedSuites.reattach(id) {
		return nil, errSuiteNotFound(r.seq, id)
	}
	h.id = id
	return h.progressHandler, nil
}

func (h *session) progressHandler(r *request) (handler, error) {
	// TODO
	return nil, nil
}

//import (
//	"context"
//	"github.com/tmazeika/testpass/repo"
//)
//
//type session struct {
//	ctx   context.Context
//	repos repo.Repos
//	suite *repo.Suite
//}
//
//type startSessionFailureType struct {
//	name        string
//	description string
//}
//
//type startSessionEnvVar struct {
//	key   string
//	value interface{}
//}
//
//type startSessionOpts struct {
//	name         string
//	failureTypes []startSessionFailureType
//	tags         []string
//	envVars      []startSessionEnvVar
//	plannedCases int64
//	startedAt    int64
//}
//
//func newSession(ctx context.Context, repos repo.Repos, opts *startSessionOpts) (*session, error) {
//	suite := repo.UnsavedSuite{
//		Name:         opts.name,
//		Tags:         opts.tags,
//		PlannedCases: opts.plannedCases,
//		Status:       repo.SuiteStatusRunning,
//		StartedAt:    opts.startedAt,
//	}
//	for _, v := range opts.failureTypes {
//		suite.FailureTypes = append(suite.FailureTypes, repo.SuiteFailureType{
//			Name:        v.name,
//			Description: v.description,
//		})
//	}
//	for _, v := range opts.envVars {
//		suite.EnvVars = append(suite.EnvVars, repo.SuiteEnvVar{
//			Key:   v.key,
//			Value: v.value,
//		})
//	}
//	ctx, cancel := context.WithTimeout(ctx, )
//	id, err := repos.Suites().Save(ctx, suite)
//	if err != nil {
//		return nil, err
//	}
//
//	return &session{
//		ctx:   ctx,
//		repos: repos,
//		suite: &repo.Suite{
//			SavedEntity: repo.SavedEntity{
//				Id: id,
//			},
//			UnsavedSuite: suite,
//		},
//	}, nil
//}
//
//func reconnectSession(ctx context.Context, repos repo.Repos, old *session) (*session, error) {
//	suite := repos.Suites().Find(ctx, )
//}
