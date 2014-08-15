package main

import (
	"errors"
	"gopkg.in/yaml.v1"
	"io/ioutil"
)

type Conf struct {
	Addr    string      `yaml:"address"`
	SSL     SSLConf     `yaml:"ssl"`
	Auth    AuthConf    `yaml:"auth"`
	Domain  []string    `yaml:"domain"`
	Proxies []ProxyConf `yaml:"proxy"`
	Htdocs  string      `yaml:"htdocs"`
}

type SSLConf struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

type AuthConf struct {
	Session AuthSessionConf `yaml:"session"`
	Info    AuthInfoConf    `yaml:"info"`
}

type AuthSessionConf struct {
	Key string `yaml:"key"`
}

type AuthInfoConf struct {
	Service      string `yaml:"service"`
	ClientId     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURL  string `yaml:"redirect_url"`
}

type ProxyConf struct {
	Path  string `yaml:"path"`
	Dest  string `yaml:"dest"`
	Strip bool   `yaml:"strip_path"`
}

func ParseConf(path string) (*Conf, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	c := &Conf{}
	if err := yaml.Unmarshal(data, c); err != nil {
		return nil, err
	}

	if c.Addr == "" {
		return nil, errors.New("address config is required")
	}

	if c.Auth.Session.Key == "" {
		return nil, errors.New("auth.session.key config is required")
	}
	if c.Auth.Info.Service == "" {
		return nil, errors.New("auth.info.service config is required")
	}
	if c.Auth.Info.ClientId == "" {
		return nil, errors.New("auth.info.client_id config is required")
	}
	if c.Auth.Info.ClientSecret == "" {
		return nil, errors.New("auth.info.client_secret config is required")
	}
	if c.Auth.Info.RedirectURL == "" {
		return nil, errors.New("auth.info.redirect_url config is required")
	}

	if c.Htdocs == "" {
		c.Htdocs = "."
	}

	return c, nil
}
