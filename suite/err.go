package suite

const (
	typeError         responseType = "error"
	typeBadSeq        responseType = "bad_seq"
	typeBadCmd        responseType = "bad_cmd"
	typeBadSuite      responseType = "bad_suite"
	typeSuiteNotFound responseType = "suite_not_found"
)

func errOther(seq int64, err error) error {
	return &response{
		Type: typeError,
		Seq:  seq,
		Payload: map[string]interface{}{
			"reason": err.Error(),
		},
	}
}

func errBadSeq(seq interface{}, reason string) error {
	return &response{
		Type: typeBadSeq,
		Payload: map[string]interface{}{
			"seq":    seq,
			"reason": reason,
		},
	}
}

func errBadCmd(seq int64, cmd interface{}, reason string) error {
	return &response{
		Type: typeBadCmd,
		Seq:  seq,
		Payload: map[string]interface{}{
			"cmd":    cmd,
			"reason": reason,
		},
	}
}

func errBadSuite(seq int64, suite interface{}, reason string) error {
	return &response{
		Type: typeBadSuite,
		Seq:  seq,
		Payload: map[string]interface{}{
			"suite":  suite,
			"reason": reason,
		},
	}
}

func errSuiteNotFound(seq int64, id interface{}) error {
	return &response{
		Type: typeSuiteNotFound,
		Seq:  seq,
		Payload: map[string]interface{}{
			"id": id,
		},
	}
}
