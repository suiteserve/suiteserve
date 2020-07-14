package api

import (
	"context"
	"github.com/gorilla/websocket"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

const (
	protoV1         = "suiteserve/1"
	shutdownTimeout = 3 * time.Second
)

type Options struct {
	Host      string
	Port      string
	CertFile  string
	KeyFile   string
	PublicDir string

	UserContentHost     string
	UserContentDir      string
	UserContentMetaRepo UserContentMetaRepo
}

type Server struct {
	cancel   func()
	opts     Options
	rpc      *rpc.Server
	srv      http.Server
	stopping bool
	stopOnce sync.Once
	wg       sync.WaitGroup
}

func newServer(opts Options) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	var s Server
	s.cancel = cancel
	s.opts = opts
	s.rpc = rpc.NewServer()
	s.srv.Addr = net.JoinHostPort(opts.Host, opts.Port)
	s.srv.BaseContext = func(net.Listener) context.Context {
		return ctx
	}
	s.setHandler()
	s.initRpc()
	return &s
}

func Serve(opts Options) *Server {
	s := newServer(opts)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.serve()
	}()
	return s
}

func (s *Server) Stop() {
	s.stopOnce.Do(s.stop)
}

func (s *Server) serve() {
	log.Printf("Binding HTTP to %s", net.JoinHostPort(s.opts.Host, s.opts.Port))
	err := s.srv.ListenAndServeTLS(s.opts.CertFile, s.opts.KeyFile)
	if err != http.ErrServerClosed {
		log.Fatalf("listen and serve http: %v", err)
	}
}

func (s *Server) stop() {
	log.Print("Shutting down HTTP...")
	s.stopping = true
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		log.Printf("close http: %v", err)
	}
	s.cancel()
	s.wg.Wait()
}

func (s *Server) setHandler() {
	var mux http.ServeMux
	mux.Handle("/changefeed",
		newHijackMiddleware(s,
			s.newWsHandler()))
	mux.Handle("/instance",
		newHijackMiddleware(s,
			s.newTcpHandler()))
	mux.Handle("/",
		newSecurityMiddleware(
			newFrontendSecurityMiddleware(
				http.FileServer(http.Dir(s.opts.PublicDir)))))
	mux.Handle(s.opts.UserContentHost+"/",
		newSecurityMiddleware(
			newUserContentSecurityMiddleware(
				newUserContentMiddleware(s.opts.UserContentMetaRepo,
					http.FileServer(http.Dir(s.opts.UserContentDir))))))
	s.srv.Handler = newLoggingMiddleware(newGetMiddleware(&mux))
}

type wsClient struct {
	conn *websocket.Conn
	rmu  sync.Mutex
	wmu  sync.Mutex
}

func (c *wsClient) Read(p []byte) (int, error) {
	c.rmu.Lock()
	defer c.rmu.Unlock()
	_, r, err := c.conn.NextReader()
	if err != nil {
		return 0, err
	}
	return r.Read(p)
}

func (c *wsClient) Write(p []byte) (int, error) {
	c.wmu.Lock()
	defer c.wmu.Unlock()
	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return 0, err
	}
	n, err := w.Write(p)
	if err != nil {
		return n, err
	}
	return n, w.Close()
}

func (c *wsClient) Close() error {
	return c.conn.Close()
}

func (s *Server) newWsHandler() http.Handler {
	upgrader := websocket.Upgrader{
		Subprotocols: []string{protoV1},
		CheckOrigin: func(r *http.Request) bool {
			// TODO: not for production; may want a 'dev' config option
			return true
		},
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("upgrade ws: %v", err)
			return
		}
		go func() {
			<-r.Context().Done()
			if err := conn.Close(); err != nil {
				log.Printf("close ws: %v", err)
			}
		}()
		s.serveRpc(&wsClient{conn: conn})
	})
}

func (s *Server) newTcpHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: respond
		conn, rw, err := w.(http.Hijacker).Hijack()
		if err != nil {
			log.Printf("upgrade tcp: %v", err)
			return
		}
		go func() {
			<-r.Context().Done()
			if err := conn.Close(); err != nil {
				log.Printf("close tcp: %v", err)
			}
		}()
		if err := conn.SetDeadline(time.Time{}); err != nil {
			log.Printf("set tcp deadline: %v", err)
			return
		}
		if _, err := io.Copy(ioutil.Discard, rw); err != nil {
			log.Printf("discard tcp rw: %v", err)
		}
		s.serveRpc(conn)
	})
}
