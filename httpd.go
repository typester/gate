package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"encoding/base64"
	"encoding/json"

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

func NewServer(conf *Conf) *Server {
	return &Server{conf}
}

func (s *Server) Run() error {
	m := martini.Classic()

	m.Use(sessions.Sessions("session", sessions.NewCookieStore([]byte(s.Conf.Auth.Session.Key))))
	m.Use(oauth2.Google(&oauth2.Options{
		ClientId:     s.Conf.Auth.Google.ClientId,
		ClientSecret: s.Conf.Auth.Google.ClientSecret,
		RedirectURL:  s.Conf.Auth.Google.RedirectURL,
		Scopes:       []string{"email"},
	}))

	m.Use(oauth2.LoginRequired)
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
			m.Any(p.Path, http.StripPrefix(strip_path, proxy))
		} else {
			m.Any(p.Path, proxy)
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
	return func(c martini.Context, tokens oauth2.Tokens, w http.ResponseWriter) {
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
