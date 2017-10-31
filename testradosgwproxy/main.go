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

	"github.com/BurntSushi/toml"
	rgw "github.com/myENA/radosgwadmin"
	"github.com/myENA/restclient"
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

	cfg := &config{}
	_, err = toml.Decode(string(cfgFile), cfg)

	if err != nil {
		log.Fatalf("Could not parse config file: %s", err)
	}

	if cfg.Server == nil || cfg.RGW == nil {
		log.Fatalf("Need to specify both [server] and [rgw] sections of config")
	}

	target, err := url.Parse(cfg.RGW.ServerURL)
	if err != nil {
		log.Fatalf("Could not parse URL: %s", cfg.RGW.ServerURL)
	}

	aa, err := rgw.NewAdminAPI(cfg.RGW)
	if err != nil {
		log.Fatalf("Could not initialize admin api: %s", err)
	}

	p := newProxy(target, cfg, aa.Client.Client.Transport)
	http.HandleFunc("/", p.proxy.ServeHTTP)
	http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Server.ServiceHost, cfg.Server.ServicePort), nil)

}

type proxy struct {
	target *url.URL
	proxy  *httputil.ReverseProxy
	cfg    *config
	creds  awsauth.Credentials
}

func newProxy(target *url.URL, cfg *config, transport http.RoundTripper) *proxy {
	p := new(proxy)
	p.target = target
	p.cfg = cfg
	p.creds = awsauth.Credentials{
		AccessKeyID:     cfg.RGW.AccessKeyID,
		SecretAccessKey: cfg.RGW.SecretAccessKey,
		SecurityToken:   cfg.RGW.SecurityToken,
		Expiration:      cfg.RGW.Expiration,
	}
	p.proxy = new(httputil.ReverseProxy)
	p.proxy.Transport = transport
	p.proxy.Director = p.director
	return p
}

func (p *proxy) director(req *http.Request) {
	target := p.target
	targetQuery := target.RawQuery
	req.URL.Scheme = target.Scheme
	req.URL.Host = p.target.Host
	req.Host = target.Host
	req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
	req.Header.Set("Host", target.Host)
	if p.cfg.Server.UseV4Auth {
		_ = awsauth.Sign4(req, p.creds)
	} else {
		_ = awsauth.SignS3(req, p.creds)
	}
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

type config struct {
	Server *serverConfig
	RGW    *rgw.Config
}

type serverConfig struct {
	ServiceHost  string
	ServicePort  int
	ReadTimeout  restclient.Duration
	WriteTimeout restclient.Duration
	UseV4Auth    bool
}

func init() {
	flag.BoolVar(&hepMe, "help", false, "get help")
	flag.StringVar(&configFilePath, "config", "config/config.toml", "location of config toml file")
}
