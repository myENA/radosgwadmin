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

var tz *time.Location = nil

var falsch bool = false

func init() {
	tz, _ = time.LoadLocation("Local") // Defaults to local
}

func SetTimezone(loc *time.Location) {
	tz = loc
}

type FormatReq struct {
	Format string `url:"format"`
}

var frj = &FormatReq{"json"}

// Use this type whenever you want to make an API call where a bool defaults to
// true if omitted, and want it to actually be false.  In such cases, the struct
// contains a reference to a bool versus a bool, since you cannot otherwise
// differentiate false with unspecified.  This represents a reference to a
// boolean false value.
var FalseRef *bool = &falsch

type AdminApi struct {
	c     *http.Client
	u     *url.URL
	t     *http.Transport
	creds *awsauth.Credentials
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

	aa.t = new(http.Transport)
	tlsc := new(tls.Config)
	tlsc.InsecureSkipVerify = cfg.InsecureSkipVerify

	var cacert []byte
	if cfg.CACertBundlePath != "" {
		cacert, err = ioutil.ReadFile(cfg.CACertBundlePath)
		if err != nil {
			return nil, fmt.Errorf("Cannot open ca cert bundle %s: %s", cfg.CACertBundlePath, err)
		}
	}

	if len(cacert) != 0 {
		bundle := x509.NewCertPool()
		ok := bundle.AppendCertsFromPEM(cacert)
		if !ok {
			return nil, fmt.Errorf("Invalid cert bundle")
		}
		tlsc.RootCAs = bundle
		tlsc.BuildNameToCertificate()
	}

	aa.t.TLSClientConfig = tlsc
	aa.c = &http.Client{
		Timeout:   cfg.ClientTimeout.Duration,
		Transport: aa.t,
	}
	aa.creds = &cfg.Credentials
	if cfg.ZoneName != "" && tz.String() != cfg.ZoneName {
		tz, err = time.LoadLocation(cfg.ZoneName)
	}
	if err != nil {
		return nil, err
	}

	return aa, nil
}

func (aa *AdminApi) Get(path string, queryStruct interface{}, responseBody interface{}) error {
	return aa.Req("get", path, queryStruct, nil, responseBody)
}

func (aa *AdminApi) Delete(path string, queryStruct interface{}, responseBody interface{}) error {
	return aa.Req("delete", path, queryStruct, nil, responseBody)
}

func (aa *AdminApi) Post(path string, queryStruct, requestBody interface{}, responseBody interface{}) error {
	return aa.Req("post", path, queryStruct, requestBody, responseBody)
}

func (aa *AdminApi) Put(path string, queryStruct, requestBody interface{}, responseBody interface{}) error {
	return aa.Req("put", path, queryStruct, requestBody, responseBody)
}

func (aa *AdminApi) Req(verb, path string, queryStruct, requestBody, responseBody interface{}) error {
	path = strings.TrimLeft(path, "/")
	url := aa.u.String() + "/" + path

	s := sling.New().Client(aa.c).QueryStruct(frj).QueryStruct(queryStruct).BodyJSON(requestBody)

	switch verb {
	case "get":
		s = s.Get(url)
	case "post":
		s = s.Post(url)
	case "put":
		s = s.Put(url)
	case "delete":
		s = s.Delete(url)
	default:
		return fmt.Errorf("unsupported verb %s", verb)
	}

	req, err := s.Request()
	if err != nil {
		return err
	}

	_ = awsauth.Sign4(req, *aa.creds)

	// This is to appease AWS signature algorithm.  spaces must
	// be %20, go defaults to +
	req.URL.RawQuery = strings.Replace(req.URL.RawQuery, "+", "%20", -1)

	fmt.Printf("URL is: %#v\n", req.URL)

	resp, err := aa.c.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Invalid status code %d : %s : body: %s", resp.StatusCode, resp.Status, string(body))
	}
	if responseBody == nil {
		return nil
	}

	d := json.NewDecoder(resp.Body)
	return d.Decode(responseBody)
}

type Config struct {
	ClientTimeout      Duration
	ServerURL          string
	AdminPath          string
	CACertBundlePath   string
	InsecureSkipVerify bool
	ZoneName           string
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

func (aa *AdminApi) HttpClient() *http.Client {
	return aa.c
}
