package main

import (
	"flag"
	"log"
	"net/http"
	"os"

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

var httpPort = flag.String("http", ":8080", "the address to listen on ('[ip]:port') for http requests")
var httpsPort = flag.String("https", ":8443", "the address to listen on ('[ip]:port') for https requests")

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

	var m letsencrypt.Manager
	if cfg.CacheFile != "" {
		m.CacheFile(cfg.CacheFile)
	}
	if cfg.Email != "" {
		err := m.Register(cfg.Email, nil)
		if err != nil {
			log.Printf("Error registering: %s", err)
		}
	}

	http.Handle("/", cfg)

	go redirectHTTP()
	log.Fatal(vpki.ListenAndServeTLS(*httpsPort, nil, &m))
}
