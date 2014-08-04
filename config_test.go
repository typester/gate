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

  google:
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

func TestParseSingleDomain(t *testing.T) {
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

  google:
    client_id: 'secret client id'
    client_secret: 'secret client secret'
    redirect_url: 'http://example.com/oauth2callback'

htdocs: ./

proxy:
  - path: /foo
    dest: http://example.com/bar
    strip_path: yes

domain: 'example1.com'

`
	if err := ioutil.WriteFile(f.Name(), []byte(data), 0644); err != nil {
		t.Error(err)
	}

	conf, err := ParseConf(f.Name())
	if err != nil {
		t.Error(err)
	}

	if conf.Domain != "example1.com" {
		t.Errorf("unexpected address: %s", conf.Domain)
	}
}

