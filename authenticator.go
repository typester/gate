package main

import (
	"encoding/json"
	"github.com/go-martini/martini"
	gooauth2 "github.com/golang/oauth2"
	"github.com/martini-contrib/oauth2"
	"log"
	"net/http"
	"strings"
)

type Authenticator interface {
	Authenticate([]string, martini.Context, oauth2.Tokens, http.ResponseWriter, *http.Request)
}

func NewAuthenticator(conf *Conf) Authenticator {
	var a Authenticator

	if conf.Auth.Info.Service == "google" {
		oauth := oauth2.Google(&gooauth2.Options{
			ClientID:     conf.Auth.Info.ClientId,
			ClientSecret: conf.Auth.Info.ClientSecret,
			RedirectURL:  conf.Auth.Info.RedirectURL,
			Scopes:       []string{"email"},
		})
		a = &GoogleAuth{oauth}
	} else if conf.Auth.Info.Service == "github" {
		oauth := oauth2.Github(&gooauth2.Options{
			ClientID:     conf.Auth.Info.ClientId,
			ClientSecret: conf.Auth.Info.ClientSecret,
			RedirectURL:  conf.Auth.Info.RedirectURL,
			Scopes:       []string{"user:email"},
		})
		a = &GitHubAuth{oauth}
	} else {
		panic("unsupported authentication method")
	}

	return a
}

type GoogleAuth struct {
	martini.Handler
}

func (a *GoogleAuth) Authenticate(domain []string, c martini.Context, tokens oauth2.Tokens, w http.ResponseWriter, r *http.Request) {
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
		var user *User
		if len(domain) > 0 {
			for _, d := range domain {
				if strings.Contains(d, "@") {
					if d == email {
						user = &User{email}
					}
				} else {
					if strings.HasSuffix(email, "@"+d) {
						user = &User{email}
						break
					}
				}
			}
		} else {
			user = &User{email}
		}

		if user != nil {
			log.Printf("user %s logged in", email)
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

type GitHubAuth struct {
	martini.Handler
}

func (a *GitHubAuth) Authenticate(domain []string, c martini.Context, tokens oauth2.Tokens, w http.ResponseWriter, r *http.Request) {

}

func forbidden(w http.ResponseWriter) {
	w.WriteHeader(403)
	w.Write([]byte("Access denied"))
}
