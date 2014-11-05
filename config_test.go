package main

import (
	"io/ioutil"
	"os"
	"testing"
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

func TestParseNamebasedVhosts(t *testing.T) {
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
    cookie_domain: example.com

  info:
    service: 'google'
    client_id: 'secret client id'
    client_secret: 'secret client secret'
    redirect_url: 'http://example.com/oauth2callback'

htdocs: ./

proxy:
  - path: /
    host: elasticsearch.example.com
    dest: http://127.0.0.1:9200
  - path: /
    host: influxdb.example.com
    dest: http://127.0.0.1:8086
`
	if err := ioutil.WriteFile(f.Name(), []byte(data), 0644); err != nil {
		t.Error(err)
	}

	conf, err := ParseConf(f.Name())
	if err != nil {
		t.Error(err)
	}

	if conf.Auth.Session.CookieDomain != "example.com" {
		t.Errorf("unexpected cookie_domain: %s", conf.Auth.Session.CookieDomain)
	}

	if len(conf.Proxies) != 2 {
		t.Errorf("insufficient proxy definions")
	}
	es := conf.Proxies[0]
	if es.Path != "/" || es.Host != "elasticsearch.example.com" || es.Dest != "http://127.0.0.1:9200" {
		t.Errorf("unexpected proxy[0]: %#v", es)
	}

	ifdb := conf.Proxies[1]
	if ifdb.Path != "/" || ifdb.Host != "influxdb.example.com" || ifdb.Dest != "http://127.0.0.1:8086" {
		t.Errorf("unexpected proxy[1]: %#v", ifdb)
	}
}
