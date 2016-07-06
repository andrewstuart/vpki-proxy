package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"gopkg.in/yaml.v2"
)

type config struct {
	Email     string
	CacheFile string
	UseHSTS   bool
	Endpoints []endpoint

	m map[string][]endpoint
}

type endpoint struct {
	Hostname, Backend string

	Q  map[string]string
	rp *httputil.ReverseProxy
}

func (cfg *config) rm(r *http.Request) *httputil.ReverseProxy {
	if cfgs, ok := cfg.m[r.Host]; ok {
	findBackend:
		for _, c := range cfgs {
			for k, v := range c.Q {
				if r.URL.Query().Get(k) != v {
					continue findBackend
				}
			}

			return c.rp
		}
	}

	if c, ok := cfg.m["*"]; ok && len(c) > 0 {
		return c[0].rp
	}
	return nil
}

func (cfg *config) init() error {
	cfg.m = map[string][]endpoint{}

	for _, c := range cfg.Endpoints {
		log.Printf("Generating config for %s", c.Hostname)
		url, err := url.ParseRequestURI(c.Backend)
		if err != nil {
			return err
		}
		c.rp = httputil.NewSingleHostReverseProxy(url)

		if _, ok := cfg.m[c.Hostname]; !ok {
			cfg.m[c.Hostname] = []endpoint{}
		}
		cfg.m[c.Hostname] = append(cfg.m[c.Hostname], c)
	}

	return nil
}

func (cfg *config) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := cfg.rm(r)
	if handler != nil {
		w.Header().Set("Strict-Transport-Security", "max-age=10886400; includeSubDomains; preload")
		handler.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Not Found", 404)
}

func readConfig(fname string) (*config, error) {
	bs, err := ioutil.ReadFile(fname)

	if err != nil {
		return nil, err
	}

	var cfg config

	err = yaml.Unmarshal(bs, &cfg)
	if err != nil {
		return nil, err
	}

	err = cfg.init()
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
