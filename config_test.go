package main

import (
	"io/ioutil"
	"os"
	"testing"
	"fmt"
)

func TestParse(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(f.Name())

	data := `---
address: ":9999"

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

	fmt.Printf("conf: %+v\n", conf)
}
