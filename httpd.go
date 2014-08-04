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
	"encoding/json"

	"github.com/go-martini/martini"
	gooauth2 "github.com/golang/oauth2"
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

func NewServer(conf *Conf) *Server {
	return &Server{conf}
}

func (s *Server) Run() error {
	m := martini.Classic()

	m.Use(sessions.Sessions("session", sessions.NewCookieStore([]byte(s.Conf.Auth.Session.Key))))
	m.Use(oauth2.Google(&gooauth2.Options{
		ClientID:     s.Conf.Auth.Google.ClientId,
		ClientSecret: s.Conf.Auth.Google.ClientSecret,
		RedirectURL:  s.Conf.Auth.Google.RedirectURL,
		Scopes:       []string{"email"},
	}))

	m.Use(loginRequired())
	m.Use(restrictDomain(s.Conf.Domain))

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

		proxy := httputil.NewSingleHostReverseProxy(u)
		if p.Strip {
			m.Any(p.Path, http.StripPrefix(strip_path, proxyHandleWrapper(u, proxy)))
		} else {
			m.Any(p.Path, proxyHandleWrapper(u, proxy))
		}

		log.Printf("register proxy path:%s dest:%s", strip_path, u.String())
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

func forbidden(w http.ResponseWriter) {
	w.WriteHeader(403)
	w.Write([]byte("Access denied"))
}

func isWebsocket(r *http.Request) bool {
	if strings.ToLower(r.Header.Get("Connection")) == "upgrade" &&
		strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
		return true
	} else {
		return false
	}
}

func proxyHandleWrapper(u *url.URL, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// websocket?
		if isWebsocket(r) {
			target := u.Host
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

func restrictDomain(domain string) martini.Handler {
	return func(c martini.Context, tokens oauth2.Tokens, w http.ResponseWriter, r *http.Request) {
		// skip websocket
		if isWebsocket(r) {
			return
		}

		extra := tokens.ExtraData()
		if _, ok := extra["id_token"]; ok == false {
			log.Printf("id_token not found")
			forbidden(w)
			return
		}

		keys := strings.Split(extra["id_token"], ".")
		if len(keys) < 2 {
			log.Printf("invalid id_token")
			forbidden(w)
			return
		}

		data, err := base64Decode(keys[1])
		if err != nil {
			log.Printf("failed to decode base64: %s", err.Error())
			forbidden(w)
			return
		}

		var info map[string]interface{}
		if err := json.Unmarshal(data, &info); err != nil {
			log.Printf("failed to decode json: %s", err.Error())
			forbidden(w)
			return
		}

		if email, ok := info["email"].(string); ok {
			if domain == "" || strings.HasSuffix(email, "@"+domain) {
				user := &User{email}
				c.Map(user)
			} else {
				log.Printf("email doesn't allow: %s", email)
				forbidden(w)
				return
			}
		} else {
			log.Printf("email not found")
			forbidden(w)
			return
		}
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
