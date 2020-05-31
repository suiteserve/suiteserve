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

func (h *session) hello(m *msg) (handler, error) {
	if m.Cmd != "hello" {
		return nil, errBadCmd(m.Seq, m.Cmd, `expected "hello"`)
	}
	versionJson, ok := m.Payload["version"].(json.Number)
	if !ok {
		return nil, errBadVersion(m.Seq, m.Payload["version"], "expected int")
	}
	version, err := versionJson.Int64()
	if err != nil {
		return nil, errBadVersion(m.Seq, versionJson, "expected int")
	}
	if version == 1 {
		if err := h.enc.Encode(newHelloResponse(m.Seq)); err != nil {
			return nil, err
		}
		return h.entry, nil
	}
	return nil, errBadVersion(m.Seq, version, "unsupported")
}

func (h *session) entry(m *msg) (handler, error) {
	switch m.Cmd {
	case "start_new":
		return h.startNew(m)
	case "reattach":
		return h.reattach(m)
	default:
		return nil, errBadCmd(m.Seq, m.Cmd, `expected one of ["start_new", "reattach"]`)
	}
}

func (h *session) startNew(m *msg) (handler, error) {
	suite, ok := m.Payload["suite"].(map[string]interface{})
	if !ok {
		return nil, errBadSuite(m.Seq, m.Payload["suite"], "expected object")
	}

	// TODO
	log.Println("DEBUG", suite)
	return nil, nil
}

func (h *session) reattach(m *msg) (handler, error) {
	id, ok := m.Payload["id"].(string)
	if !ok {
		return nil, errSuiteNotFound(m.Seq, m.Payload["id"])
	}
	if !h.detachedSuites.reattach(id) {
		return nil, errSuiteNotFound(m.Seq, id)
	}
	h.id = id
	return h.progressHandler, nil
}

func (h *session) progressHandler(m *msg) (handler, error) {
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
