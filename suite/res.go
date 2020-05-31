package suite

func newHelloResponse(seq int64) *msg {
	return &msg{
		Seq: seq,
		Cmd: "hello",
	}
}

func newSetSuiteIdResponse(seq int64, id string) *msg {
	return &msg{
		Seq: seq,
		Cmd: "set_suite_id",
		Payload: map[string]interface{}{
			"id": id,
		},
	}
}
