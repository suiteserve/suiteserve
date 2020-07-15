package suitesrv

import (
	"errors"
	"github.com/suiteserve/suiteserve/repo"
)

func errTmpIo(reason string) error {
	return &response{
		Cmd: "tmp_io",
		Payload: map[string]interface{}{
			"reason": reason,
		},
	}
}

func errBadJson(reason string) error {
	return &response{
		Cmd: "bad_json",
		Payload: map[string]interface{}{
			"reason": reason,
		},
	}
}

func errBadCmd(seq int64, cmd string, reason string) error {
	return &response{
		Seq: seq,
		Cmd: "bad_cmd",
		Payload: map[string]interface{}{
			"cmd":    cmd,
			"reason": reason,
		},
	}
}

func errBadPayload(seq int64, payload interface{}, err error) error {
	return &response{
		Seq: seq,
		Cmd: "bad_payload",
		Payload: map[string]interface{}{
			"payload": payload,
			"reason":  err.Error(),
		},
	}
}

func errBadVersion(seq int64, version int, reason string, supported []string) error {
	return &response{
		Seq: seq,
		Cmd: "bad_version",
		Payload: map[string]interface{}{
			"version":   version,
			"reason":    reason,
			"supported": supported,
		},
	}
}

func errSuiteNotReconnectable(seq int64, id string, err error) error {
	return &response{
		Seq: seq,
		Cmd: "suite_not_reconnectable",
		Payload: map[string]interface{}{
			"id":     id,
			"reason": err.Error(),
		},
	}
}

func isSuiteNotReconnectable(err error) bool {
	return errors.Is(err, repo.ErrNotFound) ||
		errors.Is(err, repo.ErrNotReconnectable) ||
		errors.Is(err, repo.ErrExpired)
}

func errCaseNotFound(seq int64, id string) error {
	return &response{
		Seq: seq,
		Cmd: "case_not_found",
		Payload: map[string]interface{}{
			"id": id,
		},
	}
}

func errBadStatus(seq int64, status, reason string) error {
	return &response{
		Seq: seq,
		Cmd: "bad_status",
		Payload: map[string]interface{}{
			"status": status,
			"reason": reason,
		},
	}
}

func errOther(seq int64, err error) error {
	return &response{
		Seq: seq,
		Cmd: "error",
		Payload: map[string]interface{}{
			"reason": err.Error(),
		},
	}
}
