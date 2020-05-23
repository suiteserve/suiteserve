package repo

import "github.com/tidwall/buntdb"

type buntLogRepo struct {
	*buntRepo
}

func (r *buntRepo) newLogRepo() (*buntLogRepo, error) {
	err := r.db.ReplaceIndex("logs_case", "logs:*",
		buntdb.IndexJSON("case"), IndexOptionalJSON("timestamp"), IndexOptionalJSON("seq"))
	if err != nil {
		return nil, err
	}
	return &buntLogRepo{r}, nil
}

func (r *buntLogRepo) Save(e LogEntry) (string, error) {
	return r.save(&e, LogCollection)
}

func (r *buntLogRepo) Find(id string) (*LogEntry, error) {
	var e LogEntry
	if err := r.find(LogCollection, id, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *buntLogRepo) FindAllByCase(caseId string) ([]LogEntry, error) {
	var entries []LogEntry
	err := r.findAllBy("logs_case", map[string]interface{}{"case": caseId}, &entries)
	if err != nil {
		return nil, err
	}
	return entries, nil
}
