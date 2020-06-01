package suitesrv

import (
	"context"
	"encoding/json"
	util "github.com/tmazeika/testpass/internal"
	"github.com/tmazeika/testpass/repo"
	"net"
)

type idStore []string

func (s idStore) contains(id string) bool {
	if s == nil {
		return false
	}
	for _, v := range s {
		if v == id {
			return true
		}
	}
	return false
}

func (s idStore) remove(id string) []string {
	if s == nil {
		return []string{}
	}
	for i, v := range s {
		if v == id {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

type session struct {
	*Server
	ctx     context.Context
	enc     *json.Encoder
	suiteId string
	caseIds idStore
}

func (s *Server) newSession(ctx context.Context, conn net.Conn) *session {
	return &session{
		Server: s,
		ctx:    ctx,
		enc:    json.NewEncoder(conn),
	}
}

func (s *session) disconnect() error {
	if s.suiteId == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()
	opts := repo.NewSuiteRepoSaveStatusOptions().DisconnectedAt(util.NowTimeMillis())
	err := s.repos.Suites().SaveStatus(ctx, s.suiteId, repo.SuiteStatusDisconnected, opts)
	if err != nil {
		return err
	}
	return nil
}

func (s *session) hello(r *request) (handler, error) {
	if r.Cmd != "hello" {
		return nil, errBadCmd(r.Seq, r.Cmd, `expected "hello"`)
	}

	var payload struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(r.Payload, &payload); err != nil {
		return nil, errBadPayload(r.Seq, r.Payload, err)
	}

	if payload.Version != 1 {
		return nil, errBadVersion(r.Seq, payload.Version, "unsupported")
	}
	if err := s.enc.Encode(newHelloResponse(r.Seq)); err != nil {
		return nil, err
	}
	return s.entry, nil
}

func (s *session) entry(r *request) (handler, error) {
	switch r.Cmd {
	case "new_suite":
		return s.newSuite(r)
	case "reconnect":
		return s.reconnect(r)
	default:
		return nil, errBadCmd(r.Seq, r.Cmd,
			`expected one of ["new_suite", "reconnect"]`)
	}
}

func (s *session) newSuite(r *request) (handler, error) {
	var payload struct {
		Name         string                  `json:"name"`
		FailureTypes []repo.SuiteFailureType `json:"failure_types"`
		Tags         []string                `json:"tags"`
		EnvVars      []repo.SuiteEnvVar      `json:"env_vars"`
		PlannedCases int64                   `json:"planned_cases"`
		StartedAt    int64                   `json:"started_at"`
	}
	if err := json.Unmarshal(r.Payload, &payload); err != nil {
		return nil, errBadPayload(r.Seq, r.Payload, err)
	}

	unsavedSuite := repo.UnsavedSuite{
		Name:         payload.Name,
		FailureTypes: payload.FailureTypes,
		Tags:         payload.Tags,
		EnvVars:      payload.EnvVars,
		PlannedCases: payload.PlannedCases,
		Status:       repo.SuiteStatusRunning,
		StartedAt:    payload.StartedAt,
	}
	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()
	id, err := s.repos.Suites().Save(ctx, unsavedSuite)
	if err != nil {
		return nil, errOther(r.Seq, err)
	}
	s.suiteId = id
	if err := s.enc.Encode(newCreatedResponse(r.Seq, id)); err != nil {
		return nil, err
	}
	return s.inProgress, nil
}

func (s *session) reconnect(r *request) (handler, error) {
	now := util.NowTimeMillis()
	var payload struct {
		Id string `json:"id"`
	}
	if err := json.Unmarshal(r.Payload, &payload); err != nil {
		return nil, errBadPayload(r.Seq, r.Payload, err)
	}
	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()
	err := s.repos.Suites().Reconnect(ctx, payload.Id, now, s.reconnectPeriod)
	if err == repo.ErrNotFound || err == repo.ErrNotReconnectable || err == repo.ErrExpired {
		return nil, errSuiteNotReconnectable(r.Seq, payload.Id, err)
	} else if err != nil {
		return nil, errOther(r.Seq, err)
	}

	cases, err := s.repos.Cases().FindAllBySuite(ctx, payload.Id, nil)
	if err != nil {
		return nil, errOther(r.Seq, err)
	}
	for _, c := range cases {
		s.caseIds = append(s.caseIds, c.Id)
	}

	s.suiteId = payload.Id
	if err := s.enc.Encode(newCreatedResponse(r.Seq, payload.Id)); err != nil {
		return nil, err
	}
	return s.inProgress, nil
}

func (s *session) inProgress(r *request) (handler, error) {
	switch r.Cmd {
	case "new_case":
		return s.newCase(r)
	case "set_case_status":
		return s.setCaseStatus(r)
	case "new_log_entry":
		return s.newLogEntry(r)
	case "set_suite_status":
		return s.setSuiteStatus(r)
	default:
		return nil, errBadCmd(r.Seq, r.Cmd,
			`expected one of ["new_case", "set_case_status", "new_log_entry", "set_suite_status"]`)
	}
}

func (s *session) newCase(r *request) (handler, error) {
	var payload struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Tags        []string        `json:"tags"`
		Num         int64           `json:"num"`
		Links       []repo.CaseLink `json:"links"`
		Args        []repo.CaseArg  `json:"args"`
		Disabled    bool            `json:"disabled"`
		CreatedAt   int64           `json:"created_at"`
	}
	if err := json.Unmarshal(r.Payload, &payload); err != nil {
		return nil, errBadPayload(r.Seq, r.Payload, err)
	}

	unsavedCase := repo.UnsavedCase{
		Suite:       s.suiteId,
		Name:        payload.Name,
		Description: payload.Description,
		Tags:        payload.Tags,
		Num:         payload.Num,
		Links:       payload.Links,
		Args:        payload.Args,
		Status:      repo.CaseStatusCreated,
		CreatedAt:   payload.CreatedAt,
	}
	if payload.Disabled {
		unsavedCase.Status = repo.CaseStatusDisabled
		unsavedCase.StartedAt = payload.CreatedAt
		unsavedCase.FinishedAt = payload.CreatedAt
	}
	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()
	id, err := s.repos.Cases().Save(ctx, unsavedCase)
	if err != nil {
		return nil, errOther(r.Seq, err)
	}
	if !payload.Disabled {
		s.caseIds = append(s.caseIds, id)
	}

	if err := s.enc.Encode(newCreatedResponse(r.Seq, id)); err != nil {
		return nil, err
	}
	return s.inProgress, nil
}

func (s *session) setCaseStatus(r *request) (handler, error) {
	var payload struct {
		Id     string          `json:"id"`
		Status repo.CaseStatus `json:"status"`
		At     int64           `json:"at"`
	}
	if err := json.Unmarshal(r.Payload, &payload); err != nil {
		return nil, errBadPayload(r.Seq, r.Payload, err)
	}
	if payload.Status == repo.CaseStatusCreated {
		return nil, errBadStatus(r.Seq, string(repo.CaseStatusCreated),
			"cannot go back to created status")
	}

	if !s.caseIds.contains(payload.Id) {
		return nil, errCaseNotFound(r.Seq, payload.Id)
	}

	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()
	opts := repo.NewCaseRepoSaveStatusOptions()
	if payload.Status == repo.CaseStatusRunning || payload.Status == repo.CaseStatusDisabled {
		opts.StartedAt(payload.At)
	}
	if payload.Status != repo.CaseStatusRunning {
		opts.FinishedAt(payload.At)
	}
	err := s.repos.Cases().SaveStatus(ctx, payload.Id, payload.Status, opts)
	if err != nil {
		return nil, errOther(r.Seq, err)
	}

	if payload.Status != repo.CaseStatusRunning {
		s.caseIds = s.caseIds.remove(payload.Id)
	}
	if err := s.enc.Encode(newOkResponse(r.Seq)); err != nil {
		return nil, err
	}
	return s.inProgress, nil
}

func (s *session) newLogEntry(r *request) (handler, error) {
	var payload repo.UnsavedLogEntry
	if err := json.Unmarshal(r.Payload, &payload); err != nil {
		return nil, errBadPayload(r.Seq, r.Payload, err)
	}

	if !s.caseIds.contains(payload.Case) {
		return nil, errCaseNotFound(r.Seq, payload.Case)
	}

	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()
	if _, err := s.repos.Logs().Save(ctx, payload); err != nil {
		return nil, errOther(r.Seq, err)
	}

	if err := s.enc.Encode(newOkResponse(r.Seq)); err != nil {
		return nil, err
	}
	return s.inProgress, nil
}

func (s *session) setSuiteStatus(r *request) (handler, error) {
	var payload struct {
		Status repo.SuiteStatus `json:"status"`
		At     int64            `json:"at"`
	}
	if err := json.Unmarshal(r.Payload, &payload); err != nil {
		return nil, errBadPayload(r.Seq, r.Payload, err)
	}
	if payload.Status != repo.SuiteStatusPassed && payload.Status != repo.SuiteStatusFailed {
		return nil, errBadStatus(r.Seq, string(payload.Status),
			"can only set status to passed or failed")
	}

	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()
	opts := repo.NewSuiteRepoSaveStatusOptions().FinishedAt(payload.At)
	err := s.repos.Suites().SaveStatus(ctx, s.suiteId, payload.Status, opts)
	if err != nil {
		return nil, errOther(r.Seq, err)
	}

	cases, err := s.repos.Cases().FindAllBySuite(ctx, s.suiteId, nil)
	if err != nil {
		return nil, errOther(r.Seq, err)
	}
	caseOpts := repo.NewCaseRepoSaveStatusOptions().FinishedAt(payload.At)
	for _, c := range cases {
		if !c.Finished() {
			err := s.repos.Cases().SaveStatus(ctx, c.Id, repo.CaseStatusErrored, caseOpts)
			if err != nil {
				return nil, errOther(r.Seq, err)
			}
		}
	}

	if err := s.enc.Encode(newOkResponse(r.Seq)); err != nil {
		return nil, err
	}
	return nil, nil
}
