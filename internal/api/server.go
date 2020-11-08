package api

import (
	"context"
	"encoding/json"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

const timeout = 3 * time.Second

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

func (o Options) newHandler() http.Handler {
	var m http.ServeMux
	m.Handle("/v1/",
		http.StripPrefix("/v1", o.V1))
	m.Handle(o.UserContentHost+"/",
		userContentHandler(o.UserContentRepo, o.UserContentDir))
	m.Handle("/",
		uiHandler(o.PublicDir))
	return logMw(secMw(&m))
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
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return srv.Shutdown(ctx)
	})
	return eg.Wait()
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
			code:  http.StatusBadRequest,
			cause: err,
		}
	}
	return nil
}

func writeJson(w http.ResponseWriter, r *http.Request, v interface{}) error {
	b, err := json.Marshal(v)
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
