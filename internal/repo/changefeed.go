package repo

type Mask []string

type Changefeed []Change

type Change interface {
	isChange()
}

type SuiteInsert struct {
	Suite Suite
	Agg   SuiteAgg
}

func (SuiteInsert) isChange() {}

type SuiteUpdate struct {
	SuiteInsert
	Mask
}
