package radosgwadmin

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/myENA/restclient"
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
	*restclient.BaseClient
	creds *awsauth.Credentials
}

// NewAdminAPI - AdminAPI factory method.
func NewAdminAPI(cfg *Config) (*AdminAPI, error) {
	baseURL := strings.Trim(cfg.ServerURL, "/")
	adminPath := strings.Trim(cfg.AdminPath, "/")
	aa := &AdminAPI{}
	var err error
	u, err := url.Parse(baseURL + "/" + adminPath)
	if err != nil {
		return nil, err
	}

	c, err := restclient.NewClient(&(cfg.ClientConfig), nil)
	if err != nil {
		return nil, err
	}
	aa.BaseClient = &restclient.BaseClient{Client: c, BaseURL: u}
	aa.Client.FixupCallback = aa.fixupCallback

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

// Config - this configures an AdminAPI.
//
// Specify CACertBundlePath to load a bundle from disk to override the default.
// Specify CACertBundle if you want embed the cacert bundle in PEM format.
// Specify one or the other.  If both are specified, CACertBundle is honored.
type Config struct {
	restclient.ClientConfig
	ServerURL        string
	AdminPath        string
	CACertBundlePath string
	ZoneName         string
	AccessKeyID      string
	SecretAccessKey  string
	SecurityToken    string
	Expiration       time.Time
}

func (aa *AdminAPI) fixupCallback(req *http.Request) error {
	req.URL.Query().Set("format", "json")

	_ = awsauth.SignS3(req, *aa.creds)

	// This is to appease AWS signature algorithm.  spaces must
	// be %20, go defaults to +
	req.URL.RawQuery = strings.Replace(req.URL.RawQuery, "+", "%20", -1)
	return nil
}
