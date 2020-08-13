package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/suiteserve/suiteserve/middleware"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

type Options struct {
	TlsCertFile string
	TlsKeyFile  string
	PublicDir   string

	UserContentHost string
	UserContentDir  string
	UserContentRepo FileMetaRepo

	V1 http.Handler
}

func (o Options) newHandler() http.Handler {
	var m http.ServeMux
	m.Handle("/v1/", http.StripPrefix("/v1", o.V1))
	m.Handle(o.UserContentHost+"/",
		newSecMiddleware(
			newUserContentMiddleware(o.UserContentRepo,
				http.FileServer(http.Dir(o.UserContentDir)))))
	m.Handle("/",
		newSecMiddleware(
			newUiSecMiddleware(
				http.FileServer(http.Dir(o.PublicDir)))))
	return middleware.Log(&m)
}

func Serve(ctx context.Context, addr string, opts Options) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("Listening at %s", ln.Addr())

	srv := http.Server{
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		Handler: opts.newHandler(),
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

type httpError struct {
	error string
	code  int
	cause error
}

func (e httpError) Error() string {
	if e.error == "" {
		return http.StatusText(e.code)
	}
	return e.error
}

func (e httpError) Unwrap() error {
	return e.cause
}

func errHandler(
	h func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			httpErr := httpError{
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
			log.Printf("<%s> %d: %s", r.RemoteAddr, httpErr.code,
				text)
			http.Error(w, httpErr.Error(), httpErr.code)
		}
	}
}

func readJson(r *http.Request, dst interface{}) error {
	if r.Header.Get("content-type") != "application/json" {
		return httpError{code: http.StatusUnsupportedMediaType}
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, dst); err != nil {
		return httpError{
			error: "bad json",
			code:  http.StatusBadRequest,
			cause: err,
		}
	}
	return nil
}

func writeJson(w http.ResponseWriter, r *http.Request, src interface{}) error {
	b, err := json.Marshal(src)
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
	var foundErr interface {
		Found() bool
	}
	return errors.As(err, &foundErr) && !foundErr.Found()
}
