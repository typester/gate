package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestPrepareFoo(t *testing.T) {
	http.HandleFunc("/foo/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello Foo\n")
	})
	go func() {
		err := http.ListenAndServe(":10001", nil)
		if err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(1 * time.Second)
}

func TestPrepareBar(t *testing.T) {
	http.HandleFunc("/bar/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello Bar\n")
	})
	go func() {
		err := http.ListenAndServe(":10002", nil)
		if err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(1 * time.Second)
}

func TestRunHTTPd(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	data := `
address: "127.0.0.1:9999"
auth:
  session:
    key: dummy
  info:
    service: nothing
    client_id: dummy
    client_secret: dummy
    redirect_url: "http://example.com/oauth2callback"
proxy:
  - path: /foo
    dest: http://127.0.0.1:10001
    strip_path: no

  - path: /bar
    dest: http://127.0.0.1:10002
    strip_path: no
`
	if err := ioutil.WriteFile(f.Name(), []byte(data), 0644); err != nil {
		t.Error(err)
	}
	conf, err := ParseConf(f.Name())
	if err != nil {
		t.Error(err)
	}
	server := NewServer(conf)
	if server == nil {
		t.Error("NewServer failed")
	}
	go server.Run()
	time.Sleep(1 * time.Second)

	// backend foo
	if res, err := http.Get("http://127.0.0.1:9999/foo/"); err == nil {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		if string(body) != "hello Foo\n" {
			t.Errorf("unexpected foo body %s", body)
		}
	} else {
		t.Error(err)
	}

	// backend bar
	if res, err := http.Get("http://127.0.0.1:9999/bar/"); err == nil {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		if string(body) != "hello Bar\n" {
			t.Errorf("unexpected bar body %s", body)
		}
	} else {
		t.Error(err)
	}
}

func TestRunVhost(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	data := `
address: "127.0.0.1:10000"
auth:
  session:
    key: dummy
    cookie_domain: example.com
  info:
    service: nothing
    client_id: dummy
    client_secret: dummy
    redirect_url: "http://example.com/oauth2callback"
proxy:
  - path: /
    dest: http://127.0.0.1:10001
    strip_path: no
    host: foo.example.com

  - path: /
    dest: http://127.0.0.1:10002
    strip_path: no
    host: bar.example.com
`
	if err := ioutil.WriteFile(f.Name(), []byte(data), 0644); err != nil {
		t.Error(err)
	}
	conf, err := ParseConf(f.Name())
	if err != nil {
		t.Error(err)
	}
	server := NewServer(conf)
	if server == nil {
		t.Error("NewServer failed")
	}
	go server.Run()
	time.Sleep(1 * time.Second)

	var req *http.Request
	client := &http.Client{}

	// backend foo
	req, _ = http.NewRequest("GET", "http://127.0.0.1:10000/foo/", nil)
	req.Header.Add("Host", "foo.example.com")
	if res, err := client.Do(req); err == nil {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		if string(body) != "hello Foo\n" {
			t.Errorf("unexpected foo body %s", body)
		}
	} else {
		t.Error(err)
	}

	// backend bar
	req, _ = http.NewRequest("GET", "http://127.0.0.1:10000/bar/", nil)
	req.Header.Add("Host", "bar.example.com")
	if res, err := http.Get("http://127.0.0.1:10000/bar/"); err == nil {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		if string(body) != "hello Bar\n" {
			t.Errorf("unexpected bar body %s", body)
		}
	} else {
		t.Error(err)
	}

}
