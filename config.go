package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"

	"gopkg.in/yaml.v2"
)

type config struct {
	Email     string
	CacheFile string `yaml:"cacheFile"`
	UseHSTS   bool   `yaml:"useHSTS"`
	Endpoints []endpoint

	m map[string][]endpoint
}

var (
	routeHit = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_proxy_route_hit",
		Help: "A route was found for a requested host",
	}, []string{"host"})

	routeMiss = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_proxy_route_miss",
		Help: "A route was requested that was not present.",
	}, []string{"host"})
)

func init() {
	prometheus.MustRegisterAll(routeMiss, routeHit)
}

type endpoint struct {
	Hostname, Backend, Directory string

	Q  map[string]string
	rp http.Handler
}

func (cfg *config) rm(r *http.Request) http.Handler {
	if cfgs, ok := cfg.m[r.Host]; ok {
	findBackend:
		// Check the configs for this host, validate all other selectors match
		for _, c := range cfgs {
			for k, v := range c.Q {
				if r.URL.Query().Get(k) != v {
					continue findBackend
				}
			}

			routeHit.With(prometheus.Labels{"host": r.Host}).Inc()

			return c.rp
		}
	}

	routeMiss.With(prometheus.Labels{"host": r.Host}).Inc()

	if c, ok := cfg.m["*"]; ok && len(c) > 0 {
		return c[0].rp
	}
	return nil
}

func (cfg *config) init() error {
	cfg.m = map[string][]endpoint{}

	for _, c := range cfg.Endpoints {
		switch {
		case c.Backend != "":
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
		case c.Directory != "":
			c.rp = http.FileServer(http.Dir(c.Directory))
		}
	}

	return nil
}

func (cfg *config) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Host, r.URL, r.RemoteAddr)
	handler := cfg.rm(r)
	if handler != nil {
		if cfg.UseHSTS {
			w.Header().Set("Strict-Transport-Security", "max-age=10886400; includeSubDomains; preload")
		}

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
