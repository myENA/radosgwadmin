package adminapi

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dghubble/sling"
	"github.com/smartystreets/go-aws-auth"
)

type AdminApi struct {
	c     *http.Client
	u     *url.URL
	t     *http.Transport
	creds *awsauth.Credentials // Embed that shit
}

func NewAdminApi(cfg *Config) (*AdminApi, error) {
	baseUrl := strings.Trim(cfg.ServerURL, "/")
	adminPath := strings.Trim(cfg.AdminPath, "/")
	aa := &AdminApi{}
	var err error
	aa.u, err = url.Parse(baseUrl + "/" + adminPath)
	if err != nil {
		return nil, err
	}
	cacert := ""
	if cfg.CACertBundlePath != "" {
		bytes, err := ioutil.ReadFile(cfg.CACertBundlePath)
		if err != nil {
			panic(fmt.Sprintf("Cannot open ca cert bundle %s: %s", cfg.CACertBundlePath, err))
		}
		cacert = string(bytes)
	}
	aa.t = &http.Transport{}

	var tlsc *tls.Config
	if cacert != "" {
		bundle := x509.NewCertPool()
		ok := bundle.AppendCertsFromPEM([]byte(cacert))
		if !ok {
			panic("Invalid cert bundle")
		}
		tlsc = new(tls.Config)
		tlsc.RootCAs = bundle
		tlsc.BuildNameToCertificate()
	}
	aa.t.TLSClientConfig = tlsc
	aa.c = &http.Client{
		Timeout:   cfg.ClientTimeout.Duration,
		Transport: aa.t,
	}
	aa.creds = &cfg.Credentials
	return aa, nil
}

func (aa *AdminApi) Get(path string, queryStruct interface{}, responseBody interface{}) error {
	path = strings.TrimLeft(path, "/")
	url := aa.u.String() + "/" + path
	s := sling.New().Client(aa.c).Get(url).QueryStruct(queryStruct)
	req, err := s.Request()
	if err != nil {
		return err
	}
	signed := awsauth.Sign4(req, *aa.creds)
	resp, err := aa.c.Do(signed)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Invalid status code %d : %s", resp.StatusCode, resp.Status)
	}
	d := json.NewDecoder(resp.Body)
	return d.Decode(responseBody)

}

type Config struct {
	ClientTimeout    Duration
	ServerURL        string
	AdminPath        string
	CACertBundlePath string
	awsauth.Credentials
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
