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
	"net"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/smartystreets/go-aws-auth"
	"gopkg.in/go-playground/validator.v9"
)

var tz *time.Location

var falsch = false

var validate *validator.Validate

var altMatch = regexp.MustCompile(`eq=([^=\|]+)`)

func init() {
	tz, _ = time.LoadLocation("Local") // Defaults to local
	validate = validator.New()
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
	c                  *http.Client
	u                  *url.URL
	t                  *http.Transport
	creds              *awsauth.Credentials
	rawValidatorErrors bool
}

type customDecoder interface {
	decode(data io.Reader) error
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

	aa.t = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	tlsc := new(tls.Config)
	tlsc.InsecureSkipVerify = cfg.InsecureSkipVerify

	var cacerts []byte
	if len(cfg.CACertBundle) > 0 {
		cacerts = cfg.CACertBundle
	} else if cfg.CACertBundlePath != "" {
		cacerts, err = ioutil.ReadFile(cfg.CACertBundlePath)
		if err != nil {
			return nil, fmt.Errorf("Cannot open ca cert bundle %s: %s", cfg.CACertBundlePath, err)
		}
	}

	if len(cacerts) > 0 {
		bundle := x509.NewCertPool()
		ok := bundle.AppendCertsFromPEM(cacerts)
		if !ok {
			return nil, fmt.Errorf("Invalid cert bundle")
		}
		tlsc.RootCAs = bundle
		tlsc.BuildNameToCertificate()
	}

	aa.t.TLSClientConfig = tlsc
	aa.c = &http.Client{
		Timeout:   time.Duration(cfg.ClientTimeout),
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

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Ptr:
		return v.IsNil()

	default:
		panic("Invalid interface type: " + v.Kind().String())
	}
}

func (aa *AdminAPI) req(ctx context.Context, verb, path string, queryStruct, requestBody, responseBody interface{}) error {
	path = strings.TrimLeft(path, "/")
	url := aa.u.String() + "/" + path
	if !isNil(queryStruct) {
		err := aa.validate(queryStruct)
		if err != nil {
			return err
		}
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
	if !isNil(requestBody) {
		err := aa.validate(requestBody)
		if err != nil {
			return err
		}
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

	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Invalid status code %d : %s : body: %s", resp.StatusCode, resp.Status, string(body))
	}
	if isNil(responseBody) {
		return nil
	}

	if cd, ok := responseBody.(customDecoder); ok {
		return cd.decode(resp.Body)
	}

	return json.NewDecoder(resp.Body).Decode(responseBody)
}

// make sense of the validator error types
func (aa *AdminAPI) validate(i interface{}) error {
	err := validate.Struct(i)
	if err != nil {
		if aa.rawValidatorErrors {
			return err
		}
		if verr, ok := err.(validator.ValidationErrors); ok {
			var errs []string
			for _, ferr := range verr {
				if ferr.ActualTag() == "required" {
					errs = append(errs,
						fmt.Sprintf("Required field %s is missing or empty",
							ferr.StructField(),
						),
					)
				} else if matches := altMatch.FindAllStringSubmatch(ferr.ActualTag(), -1); len(matches) > 0 {
					valids := make([]string, len(matches))
					for i := 0; i < len(matches); i++ {
						valids[i] = "\"" + matches[i][1] + "\""
					}
					errs = append(errs,
						fmt.Sprintf("Field '%s' invalid value: '%s', valid values are: %s",
							ferr.StructNamespace(),
							ferr.Value(), // for now all are string - revise this if other types are needed
							strings.Join(valids, ",")),
					)
				}
			}

			return fmt.Errorf("Validation error: %s", strings.Join(errs, " ; "))
		}
	}
	return err
}

// Config - this configures an AdminAPI.
//
// Specify CACertBundlePath to load a bundle from disk to override the default.
// Specify CACertBundle if you want embed the cacert bundle in PEMm format.
// Specify one or the other.  If both are specified, CACertBundle is honored.
type Config struct {
	ClientTimeout      Duration
	ServerURL          string
	AdminPath          string
	CACertBundlePath   string
	CACertBundle       []byte
	InsecureSkipVerify bool
	ZoneName           string
	AccessKeyID        string
	SecretAccessKey    string
	SecurityToken      string
	Expiration         time.Time
	RawValidatorErrors bool // If true, then no attempt to interpret validator errors will be made.
}

// Duration - this allows us to use a text representation of a duration and
// have it parse correctly.  The go standard library time.Duration does not
// implement the TextUnmarshaller interface, so we have to do this workaround
// in order for json.Unmarshal or external parsers like toml.Decode to work
// with human friendly input.
type Duration time.Duration

// UnmarshalText - this implements the TextUnmarshaler
func (d *Duration) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}
	dur, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

// MarshalText - this implements TextMarshaler
func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

// HTTPClient return the underlying http.Client.  You can use this to fine tune
// the http.Transport settings, for example.
func (aa *AdminAPI) HTTPClient() *http.Client {
	return aa.c
}
