package main

import (
	"flag"
	"log"
)

var (
	confFile = flag.String("conf", "config.yml", "config file path")
)

func main() {
	flag.Parse()

	conf, err := ParseConf(*confFile)
	if err != nil {
		panic(err)
	}

	server := NewServer(conf)
	log.Fatal(server.Run())
}
