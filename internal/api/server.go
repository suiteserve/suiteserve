package api

import (
	"context"
	"encoding/json"
	"errors"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

type Options struct {
	Addr        string
	TlsCertFile string
	TlsKeyFile  string
	PublicDir   string

	UserContentHost string
	UserContentDir  string
	UserContentRepo FileMetaRepo

	V1 http.Handler
}

func (o Options) handler() http.Handler {
	var m http.ServeMux
	m.Handle("/v1/",
		http.StripPrefix("/v1", o.V1))
	m.Handle(o.UserContentHost+"/",
		secMw(userContentHandler(o.UserContentRepo, o.UserContentDir)))
	m.Handle("/",
		secMw(uiHandler(o.PublicDir)))
	return logMw(&m)
}

func Serve(ctx context.Context, opts Options) error {
	ln, err := net.Listen("tcp", opts.Addr)
	if err != nil {
		return err
	}
	log.Printf("Listening at %s", ln.Addr())

	srv := http.Server{
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		Handler: opts.handler(),
	}
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		err := srv.ServeTLS(ln, opts.TlsCertFile, opts.TlsKeyFile)
		if err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	eg.Go(func() error {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	})
	return eg.Wait()
}

type errHttp struct {
	error string
	code  int
	cause error
}

func (e errHttp) Error() string {
	if e.error == "" {
		return http.StatusText(e.code)
	}
	return e.error
}

func (e errHttp) Unwrap() error {
	return e.cause
}

type errHandlerFunc func(w http.ResponseWriter, r *http.Request) error

func (f errHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := f(w, r); err != nil {
		httpErr := errHttp{
			code:  http.StatusInternalServerError,
			cause: err,
		}
		if !errors.As(err, &httpErr) && isNotFound(err) {
			httpErr.code = http.StatusNotFound
		}
		text := httpErr.Error()
		if httpErr.cause != nil {
			text += ": " + httpErr.cause.Error()
		}
		log.Printf("<%s> %d %s", r.RemoteAddr, httpErr.code, text)
		http.Error(w, httpErr.Error(), httpErr.code)
	}
}

func readJson(r *http.Request, dst interface{}) error {
	if r.Header.Get("content-type") != "application/json" {
		return errHttp{code: http.StatusUnsupportedMediaType}
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, dst); err != nil {
		return errHttp{
			error: "bad json",
			code:  http.StatusBadRequest,
			cause: err,
		}
	}
	return nil
}

func writeJson(w http.ResponseWriter, r *http.Request, x interface{}) error {
	b, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	b = append(b, '\n')
	w.Header().Set("content-length", strconv.Itoa(len(b)))
	w.Header().Set("content-type", "application/json")
	if r.Method != http.MethodHead {
		_, err = w.Write(b)
	}
	return err
}

func isNotFound(err error) bool {
	var errNotFound interface {
		NotFound() bool
	}
	return errors.As(err, &errNotFound)
}
