package repo

type errNotFound struct{}

func (errNotFound) Error() string {
	return "not found"
}

func (errNotFound) NotFound() bool {
	return true
}
