package repo

import "encoding/json"

type ChangeType string

const (
	ChangeTypeAttachment ChangeType = "attachment"
	ChangeTypeCase       ChangeType = "case"
	ChangeTypeLog        ChangeType = "log"
	ChangeTypeSuite      ChangeType = "suite"
)

type Change struct {
	Type    ChangeType      `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func newChange(coll Collection, payload json.RawMessage) *Change {
	var t ChangeType
	switch coll {
	case AttachmentColl:
		t = ChangeTypeAttachment
	case CaseColl:
		t = ChangeTypeCase
	case LogColl:
		t = ChangeTypeLog
	case SuiteColl:
		t = ChangeTypeSuite
	default:
		panic("unknown coll " + coll)
	}
	return &Change{
		Type:    t,
		Payload: payload,
	}
}
