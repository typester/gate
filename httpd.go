package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"encoding/base64"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/oauth2"
	"github.com/martini-contrib/sessions"
	"path/filepath"
)

type Server struct {
	Conf *Conf
}

type User struct {
	Email string
}

type Backend struct {
	Host      string
	URL       *url.URL
	Strip     bool
	StripPath string
}

const (
	BackendHostHeader = "X-Gate-Backend-Host"
)

func NewServer(conf *Conf) *Server {
	return &Server{conf}
}

func (s *Server) Run() error {
	m := martini.Classic()

	cookieStore := sessions.NewCookieStore([]byte(s.Conf.Auth.Session.Key))
	if domain := s.Conf.Auth.Session.CookieDomain; domain != "" {
		cookieStore.Options(sessions.Options{Domain: domain})
	}
	m.Use(sessions.Sessions("session", cookieStore))

	if s.Conf.Auth.Info.Service != noAuthServiceName {
		a := NewAuthenticator(s.Conf)
		m.Use(a.Handler())
		m.Use(loginRequired())
		m.Use(restrictRequest(s.Conf.Restrictions, a))
	}

	backendsFor := make(map[string][]Backend)
	backendIndex := make([]string, len(s.Conf.Proxies))

	for i := range s.Conf.Proxies {
		p := s.Conf.Proxies[i]

		if strings.HasSuffix(p.Path, "/") == false {
			p.Path += "/"
		}
		strip_path := p.Path

		if strings.HasSuffix(p.Path, "**") == false {
			p.Path += "**"
		}

		u, err := url.Parse(p.Dest)
		if err != nil {
			return err
		}
		backendsFor[p.Path] = append(backendsFor[p.Path], Backend{
			Host:      p.Host,
			URL:       u,
			Strip:     p.Strip,
			StripPath: strip_path,
		})
		backendIndex[i] = p.Path
		log.Printf("register proxy host:%s path:%s dest:%s strip_path:%v", p.Host, strip_path, u.String(), p.Strip)
	}

	registered := make(map[string]bool)
	for _, path := range backendIndex {
		if registered[path] {
			continue
		}
		proxy := newVirtualHostReverseProxy(backendsFor[path])
		m.Any(path, proxyHandleWrapper(proxy))
		registered[path] = true
	}

	path, err := filepath.Abs(s.Conf.Htdocs)
	if err != nil {
		return err
	}

	log.Printf("starting static file server for: %s", path)
	fileServer := http.FileServer(http.Dir(path))
	m.Get("/**", fileServer.ServeHTTP)

	log.Printf("starting server at %s", s.Conf.Addr)

	if s.Conf.SSL.Cert != "" && s.Conf.SSL.Key != "" {
		return http.ListenAndServeTLS(s.Conf.Addr, s.Conf.SSL.Cert, s.Conf.SSL.Key, m)
	} else {
		return http.ListenAndServe(s.Conf.Addr, m)
	}
}

func newVirtualHostReverseProxy(backends []Backend) http.Handler {
	bmap := make(map[string]Backend)
	for _, b := range backends {
		bmap[b.Host] = b
	}
	defaultBackend, ok := bmap[""]
	if !ok {
		defaultBackend = backends[0]
	}

	director := func(req *http.Request) {
		b, ok := bmap[req.Host]
		if !ok {
			b = defaultBackend
		}
		req.URL.Scheme = b.URL.Scheme
		req.URL.Host = b.URL.Host
		if b.Strip {
			if p := strings.TrimPrefix(req.URL.Path, b.StripPath); len(p) < len(req.URL.Path) {
				req.URL.Path = "/" + p
			}
		}
		req.Header.Set(BackendHostHeader, req.URL.Host)
		log.Println("backend url", req.URL.String())
	}
	return &httputil.ReverseProxy{Director: director}
}

func isWebsocket(r *http.Request) bool {
	if strings.ToLower(r.Header.Get("Connection")) == "upgrade" &&
		strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
		return true
	} else {
		return false
	}
}

func proxyHandleWrapper(handler http.Handler) http.Handler {
	proxy, _ := handler.(*httputil.ReverseProxy)
	director := proxy.Director

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// websocket?
		if isWebsocket(r) {
			director(r) // rewrite request headers for backend
			target := r.Header.Get(BackendHostHeader)

			if strings.HasPrefix(r.URL.Path, "/") == false {
				r.URL.Path = "/" + r.URL.Path
			}

			log.Printf("proxy ws request: %s", r.URL.String())

			// websocket proxy by bradfitz https://groups.google.com/forum/#!topic/golang-nuts/KBx9pDlvFOc
			d, err := net.Dial("tcp", target)
			if err != nil {
				http.Error(w, "Error contacting backend server.", 500)
				log.Printf("Error dialing websocket backend %s: %v", target, err)
				return
			}
			hj, ok := w.(http.Hijacker)
			if !ok {
				http.Error(w, "Not a hijacker?", 500)
				return
			}
			nc, _, err := hj.Hijack()
			if err != nil {
				log.Printf("Hijack error: %v", err)
				return
			}
			defer nc.Close()
			defer d.Close()

			err = r.Write(d)
			if err != nil {
				log.Printf("Error copying request to target: %v", err)
				return
			}

			errc := make(chan error, 2)
			cp := func(dst io.Writer, src io.Reader) {
				_, err := io.Copy(dst, src)
				errc <- err
			}
			go cp(d, nc)
			go cp(nc, d)
			<-errc
		} else {
			handler.ServeHTTP(w, r)
		}
	})
}

// base64Decode decodes the Base64url encoded string
//
// steel from code.google.com/p/goauth2/oauth/jwt
func base64Decode(s string) ([]byte, error) {
	// add back missing padding
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}

func restrictRequest(restrictions []string, authenticator Authenticator) martini.Handler {
	return func(c martini.Context, tokens oauth2.Tokens, w http.ResponseWriter, r *http.Request) {
		// skip websocket
		if isWebsocket(r) {
			return
		}

		authenticator.Authenticate(restrictions, c, tokens, w, r)
	}
}

func loginRequired() martini.Handler {
	return func(s sessions.Session, c martini.Context, w http.ResponseWriter, r *http.Request) {
		if isWebsocket(r) {
			return
		}
		c.Invoke(oauth2.LoginRequired)
	}
}
