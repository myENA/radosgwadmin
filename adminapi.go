package radosgwadmin

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/smartystreets/go-aws-auth"
)

var tz *time.Location

var falsch = false

func init() {
	tz, _ = time.LoadLocation("Local") // Defaults to local
}

// SetTimeZone - override time zone.  Not thread-safe, do
// this at initialization time or protect it with a mutex
// if necessary.
func SetTimeZone(loc *time.Location) {
	tz = loc
}

// FalseRef - Use this type whenever you want to make an API call where a bool defaults to
// true if omitted, and want it to actually be false.  In such cases, the struct
// contains a reference to a bool versus a bool, since you cannot otherwise
// differentiate false with unspecified.  This represents a reference to a
// boolean false value.
var FalseRef = &falsch

// AdminAPI - admin api struct
type AdminAPI struct {
	c     *http.Client
	u     *url.URL
	t     *http.Transport
	creds *awsauth.Credentials
}

// NewAdminAPI - AdminAPI factory method.
func NewAdminAPI(cfg *Config) (*AdminAPI, error) {
	baseURL := strings.Trim(cfg.ServerURL, "/")
	adminPath := strings.Trim(cfg.AdminPath, "/")
	aa := &AdminAPI{}
	var err error
	aa.u, err = url.Parse(baseURL + "/" + adminPath)
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
	aa.creds = &awsauth.Credentials{
		AccessKeyID:     cfg.AccessKeyID,
		SecretAccessKey: cfg.SecretAccessKey,
		SecurityToken:   cfg.SecurityToken,
		Expiration:      cfg.Expiration,
	}

	if cfg.ZoneName != "" && tz.String() != cfg.ZoneName {
		tz, err = time.LoadLocation(cfg.ZoneName)
	}
	if err != nil {
		return nil, err
	}

	return aa, nil
}

func (aa *AdminAPI) get(ctx context.Context, path string, queryStruct interface{}, responseBody interface{}) error {
	return aa.req(ctx, "GET", path, queryStruct, nil, responseBody)
}

func (aa *AdminAPI) delete(ctx context.Context, path string, queryStruct interface{}, responseBody interface{}) error {
	return aa.req(ctx, "DELETE", path, queryStruct, nil, responseBody)
}

func (aa *AdminAPI) post(ctx context.Context, path string, queryStruct, requestBody interface{}, responseBody interface{}) error {

	return aa.req(ctx, "POST", path, queryStruct, requestBody, responseBody)
}

func (aa *AdminAPI) put(ctx context.Context, path string, queryStruct, requestBody interface{}, responseBody interface{}) error {
	return aa.req(ctx, "PUT", path, queryStruct, requestBody, responseBody)
}

func (aa *AdminAPI) req(ctx context.Context, verb, path string, queryStruct, requestBody, responseBody interface{}) error {
	path = strings.TrimLeft(path, "/")
	url := aa.u.String() + "/" + path
	if queryStruct != nil {
		v, err := query.Values(queryStruct)
		if err != nil {
			return err
		}
		qs := v.Encode()
		if qs != "" {
			if strings.Contains(url, "?") {
				url = url + "&" + qs
			} else {
				url = url + "?" + qs
			}
		}
	}

	var bodyReader io.Reader
	if requestBody != nil {
		bjson, err := json.Marshal(requestBody)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(bjson)
	}
	req, err := http.NewRequest(verb, url, bodyReader)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)
	req.URL.Query().Set("format", "json")

	_ = awsauth.SignS3(req, *aa.creds)

	// This is to appease AWS signature algorithm.  spaces must
	// be %20, go defaults to +
	req.URL.RawQuery = strings.Replace(req.URL.RawQuery, "+", "%20", -1)

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

// Config - this configures an AdminAPI.
//
// Specify CACertBundlePath only if you want to override the system default
// CA cert bundle.
type Config struct {
	ClientTimeout      Duration
	ServerURL          string
	AdminPath          string
	CACertBundlePath   string
	InsecureSkipVerify bool
	ZoneName           string
	AccessKeyID        string
	SecretAccessKey    string
	SecurityToken      string
	Expiration         time.Time
}

// Duration - this allows us to use a text representation
// of a duration and have it parse correctly.  Used for
// BurntSushi toml decoder, since they didn't see fit to
// handle built-in time.Duration type for some reason.
type Duration struct {
	time.Duration
}

// UnmarshalText - this implements the TextUnmarshaller
func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

// HTTPClient return the underlying http.Client
func (aa *AdminAPI) HTTPClient() *http.Client {
	return aa.c
}
