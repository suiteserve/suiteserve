package api

// type input struct {
// 	vars map[string]string
// 	errs []string
// }
//
// func (i *input) repoId(k string, id *repo.Id) {
// 	i.withVar(k, func(v string) (err error) {
// 		*id, err = repo.ParseId(v)
// 		return
// 	})
// }
//
// func (i *input) repoTime(k string, t *repo.Time) {
// 	i.withVar(k, func(v string) (err error) {
// 		*t, err = repo.ParseTime(v)
// 		return
// 	})
// }
//
// func (i *input) repoSuiteResult(k string, res *repo.SuiteResult) {
// 	i.withVar(k, func(v string) (err error) {
// 		*res, err = repo.ParseSuiteResult(v)
// 		return
// 	})
// }
//
// func (i *input) withVar(k string, fn func(v string) error) {
// 	v, ok := i.vars[k]
// 	if !ok {
// 		panic(fmt.Sprintf("var %q not found", k))
// 	}
// 	if err := fn(v); err != nil {
// 		i.errs = append(i.errs, err.Error())
// 	}
// }
//
// func parseInput(r *http.Request, fn func(i *input)) error {
// 	i := input{vars: mux.Vars(r)}
// 	fn(&i)
// 	if len(i.errs) == 0 {
// 		return errHttp{
// 			error: strings.Join(i.errs, "\n"),
// 			code:  http.StatusBadRequest,
// 		}
// 	}
// 	return nil
// }
