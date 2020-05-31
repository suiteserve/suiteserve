package suite

func errTmpIo(reason string) error {
	return &msg{
		Cmd: "tmp_io",
		Payload: map[string]interface{}{
			"reason": reason,
		},
	}
}

func errBadJson(reason string) error {
	return &msg{
		Cmd: "bad_json",
		Payload: map[string]interface{}{
			"reason": reason,
		},
	}
}

func errBadCmd(seq int64, cmd string, reason string) error {
	return &msg{
		Seq: seq,
		Cmd: "bad_cmd",
		Payload: map[string]interface{}{
			"cmd":    cmd,
			"reason": reason,
		},
	}
}

func errBadVersion(seq int64, version interface{}, reason string) error {
	return &msg{
		Seq: seq,
		Cmd: "bad_version",
		Payload: map[string]interface{}{
			"version": version,
			"reason":  reason,
		},
	}
}

func errBadSuite(seq int64, suite interface{}, reason string) error {
	return &msg{
		Seq: seq,
		Cmd: "bad_suite",
		Payload: map[string]interface{}{
			"suite":  suite,
			"reason": reason,
		},
	}
}

func errSuiteNotFound(seq int64, id interface{}) error {
	return &msg{
		Seq: seq,
		Cmd: "suite_not_found",
		Payload: map[string]interface{}{
			"id": id,
		},
	}
}

func errOther(seq int64, err error) error {
	return &msg{
		Seq: seq,
		Cmd: "error",
		Payload: map[string]interface{}{
			"reason": err.Error(),
		},
	}
}
