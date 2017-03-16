package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/BurntSushi/toml"
	"github.com/nathanejohnson/radosgwadmin/adminapi"
)

func main() {
	flag.Parse()
	if hepMe {
		flag.PrintDefaults()
		return
	}

	cfgFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatalf("Got error opening config file: %s", err)
	}

	cfg := &Config{}
	_, err = toml.Decode(string(cfgFile), cfg)

	if err != nil {
		log.Fatalf("Could not parse config file: %s", err)
	}

	hmux := NewProxyHandler(cfg)
	hmux.HandleFunc("/", hmux.DoEverything)

	h := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.ServiceHost, cfg.Server.ServicePort),
		Handler:      hmux,
		ReadTimeout:  cfg.Server.ReadTimeout.Duration,
		WriteTimeout: cfg.Server.WriteTimeout.Duration,
	}

	err = h.ListenAndServe()
	if err != nil {
		log.Fatal("Error on http server: %s", err)
	}

}

var (
	hepMe          bool
	configFilePath string
)

type Config struct {
	Server   *ServerConfig
	AdminApi *adminapi.Config
}

type ServerConfig struct {
	ServiceHost  string
	ServicePort  int
	ReadTimeout  adminapi.Duration
	WriteTimeout adminapi.Duration
}

func init() {
	flag.BoolVar(&hepMe, "help", false, "get help")
	flag.StringVar(&configFilePath, "config", "config/config.toml", "location of config toml file")
}

type ProxyHandler struct {
	ServeMux *http.ServeMux
	AdminApi *adminapi.AdminApi
}

func NewProxyHandler(cfg *Config) *ProxyHandler {
	ph := &ProxyHandler{}
	var err error
	ph.AdminApi, err = adminapi.NewAdminApi(cfg.AdminApi)
	if err != nil {
		log.Fatalf("error creating AdminApi instance: %s", err)
	}
	ph.ServeMux = &http.ServeMux{}

	return ph
}

type resp struct {
	URL    *url.URL
	Header http.Header
	Err    string
}

func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ph.ServeMux.ServeHTTP(w, r)
}

// Handle registers the handler for the given pattern. If a handler already exists for pattern, Handle panics.
func (ph *ProxyHandler) Handle(pattern string, handler http.Handler) {
	ph.ServeMux.Handle(pattern, handler)
}

func (ph *ProxyHandler) HandleFunc(pattern string, handler http.HandlerFunc) {
	ph.ServeMux.HandleFunc(pattern, handler)
}

func (ph *ProxyHandler) DoEverything(w http.ResponseWriter, req *http.Request) {
	resp, err := ph.AdminApi.GetUsage(&adminapi.UsageRequest{})
	w.Header().Set("Content-Type", "application/json")
	je := json.NewEncoder(w)

	if err != nil {
		log.Printf("Shit, got error: %s", err)
		w.WriteHeader(500)
		errMap := make(map[string]string)
		errMap["error"] = err.Error()
		je.Encode(errMap)
		return
	}

	err = je.Encode(resp)
	if err != nil {
		log.Printf("Shit, got error: %s", err)

	}

}
