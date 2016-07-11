package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"astuart.co/vpki"

	"rsc.io/letsencrypt"
)

func redirectHTTP() {
	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+r.Host+"/"+r.URL.String(), http.StatusFound)
	})
	log.Fatal(http.ListenAndServe(*httpPort, m))
}

var (
	httpPort  = flag.String("http", ":8080", "the address to listen on ('[ip]:port') for http requests")
	httpsPort = flag.String("https", ":8443", "the address to listen on ('[ip]:port') for https requests")
	metricIP  = flag.String("metric-ip", "", "an IP address for which metrics may be sent back")
)

func init() {
	flag.Parse()
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <config-file>", os.Args[0])
	}

	cfg, err := readConfig(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Config: %#v\n", cfg)

	var m letsencrypt.Manager
	if cfg.CacheFile != "" {
		log.Println("Using cache file", cfg.CacheFile)
		m.CacheFile(cfg.CacheFile)
	}
	if cfg.Email != "" {
		err := m.Register(cfg.Email, nil)
		if err != nil {
			log.Printf("Error registering: %s", err)
		}
	}

	http.Handle("/", prometheus.InstrumentHandler("proxy", cfg))

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if *metricIP != "" {
			ra := strings.Split(r.RemoteAddr, ":")
			if len(ra) < 2 {
				log.Println("Unexpected remote address without port")
				return
			}

			if ra[0] != *metricIP {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}

		prometheus.Handler().ServeHTTP(w, r)
	})

	go redirectHTTP()
	log.Fatal(vpki.ListenAndServeTLS(*httpsPort, nil, &m))
}
