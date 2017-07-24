// This is for proxying to a radosgw api instance and testing endpoints,
// such as with Postman etc.  This is a testing tool.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"bitbucket.ena.net/go/radosgwadmin/adminapi"
	"github.com/BurntSushi/toml"
	"github.com/smartystreets/go-aws-auth"
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

	if cfg.Server == nil || cfg.AdminAPI == nil {
		log.Fatalf("Need to specify both [server] and [adminapi] sections of config")
	}

	target, err := url.Parse(cfg.AdminAPI.ServerURL)
	if err != nil {
		log.Fatalf("Could not parse URL: %s", cfg.AdminAPI.ServerURL)
	}

	aa, err := adminapi.NewAdminAPI(cfg.AdminAPI)
	if err != nil {
		log.Fatalf("Could not initialize admin api: %s", err)
	}

	p := NewProxy(target, cfg, aa.HTTPClient().Transport)
	http.HandleFunc("/", p.proxy.ServeHTTP)
	http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Server.ServiceHost, cfg.Server.ServicePort), nil)

}

type Proxy struct {
	target *url.URL
	proxy  *httputil.ReverseProxy
	cfg    *Config
}

func NewProxy(target *url.URL, cfg *Config, transport http.RoundTripper) *Proxy {
	p := new(Proxy)
	p.target = target
	p.cfg = cfg
	p.proxy = new(httputil.ReverseProxy)
	p.proxy.Transport = transport
	p.proxy.Director = p.Director
	return p
}

func (p *Proxy) Director(req *http.Request) {
	target := p.target
	targetQuery := target.RawQuery
	req.URL.Scheme = target.Scheme
	req.URL.Host = p.target.Host
	req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
	// _ = awsauth.Sign4(req, p.cfg.AdminAPI.Credentials)
	_ = awsauth.SignS3(req, p.cfg.AdminAPI.Credentials)
	req.Header.Set("Host", target.Host)
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
	req.URL.RawQuery = strings.Replace(req.URL.RawQuery, "+", "%20", -1)
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

var (
	hepMe          bool
	configFilePath string
)

type Config struct {
	Server   *ServerConfig
	AdminAPI *adminapi.Config
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
