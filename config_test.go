package main

import (
	"io/ioutil"
	"os"
	"testing"
	"github.com/martini-contrib/oauth2"
)

func TestParse(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	data := `---
address: ":9999"

auth:
  session:
    key: secret

  info:
    service: 'google'
    client_id: 'secret client id'
    client_secret: 'secret client secret'
    redirect_url: 'http://example.com/oauth2callback'

htdocs: ./

proxy:
  - path: /foo
    dest: http://example.com/bar
    strip_path: yes
`
	if err := ioutil.WriteFile(f.Name(), []byte(data), 0644); err != nil {
		t.Error(err)
	}

	conf, err := ParseConf(f.Name())
	if err != nil {
		t.Error(err)
	}

	if conf.Addr != ":9999" {
		t.Errorf("unexpected address: %s", conf.Addr)
	}
}

func TestParseMultiRestrictions(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	data := `---
address: ":9999"

auth:
  session:
    key: secret

  info:
    service: 'google'
    client_id: 'secret client id'
    client_secret: 'secret client secret'
    redirect_url: 'http://example.com/oauth2callback'

htdocs: ./

proxy:
  - path: /foo
    dest: http://example.com/bar
    strip_path: yes

restrictions:
  - 'example1.com'
  - 'example2.com'
`
	if err := ioutil.WriteFile(f.Name(), []byte(data), 0644); err != nil {
		t.Error(err)
	}

	conf, err := ParseConf(f.Name())
	if err != nil {
		t.Error(err)
	}

	if len(conf.Restrictions) != 2 {
		t.Errorf("unexpected restrictions num: %d", len(conf.Restrictions))
	}

	if conf.Restrictions[0] != "example1.com" || conf.Restrictions[1] != "example2.com" {
		t.Errorf("unexpected restrictions: %+v", conf.Restrictions)
	}
}

func TestParseGithubServiceShouldSetDefaultValue(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	data := `---
address: ":9999"

auth:
  session:
    key: secret

  info:
    service: 'github'
    client_id: 'secret client id'
    client_secret: 'secret client secret'
    redirect_url: 'http://example.com/oauth2callback'
`
	if err := ioutil.WriteFile(f.Name(), []byte(data), 0644); err != nil {
		t.Error(err)
	}

	conf, err := ParseConf(f.Name())
	if err != nil {
		t.Error(err)
	}

	if conf.Auth.Info.Endpoint != "https://github.com" {
		t.Errorf("unexpected endpoint address: %s", conf.Auth.Info.Endpoint)
	}
	if conf.Auth.Info.ApiEndpoint != "https://api.github.com" {
		t.Errorf("unexpected api endpoint address: %s", conf.Auth.Info.ApiEndpoint)
	}
}

func TestPathConf(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	data := `---
address: ":9999"

auth:
  session:
    key: secret

  info:
    service: 'github'
    client_id: 'secret client id'
    client_secret: 'secret client secret'
    redirect_url: 'http://example.com/_gate_callback'

paths:
    login: "/_gate_login"
    logout: "/_gate_logout"
    callback: "/_gate_callback"
    error: "/_gate_error"
`
	if err := ioutil.WriteFile(f.Name(), []byte(data), 0644); err != nil {
		t.Error(err)
	}

	conf, err := ParseConf(f.Name())
	if err != nil {
		t.Error(err)
	}

	conf.SetOAuth2Paths()

	if oauth2.PathLogin != "/_gate_login" {
		t.Errorf("unexpected oauth2.PathLogin: %s", oauth2.PathLogin)
	}
	if oauth2.PathLogout != "/_gate_logout" {
		t.Errorf("unexpected oauth2.PathLogout: %s", oauth2.PathLogout)
	}
	if oauth2.PathCallback != "/_gate_callback" {
		t.Errorf("unexpected oauth2.PathCallback: %s", oauth2.PathCallback)
	}
	if oauth2.PathError != "/_gate_error" {
		t.Errorf("unexpected oauth2.PathError: %s", oauth2.PathError)
	}
}
