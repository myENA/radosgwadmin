// +build integration

package radosgwadmin

import (
	"context"
	"fmt"
	// "io"
	"io/ioutil"
	"log"
	// "math/rand"
	"os"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/suite"
)

type IntegrationsSuite struct {
	suite.Suite
	aa           *AdminAPI
	randFilePath string
	lf           *os.File
}

type IntegrationConfig struct {
	Integration *Integration
	AdminAPI    *Config
}

type Integration struct {
	TestUID string
}

func (is *IntegrationsSuite) SetupSuite() {

	logPath := os.Getenv("ADMINAPI_TEST_LOGFILE")
	if logPath != "" {
		lf, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(fmt.Sprintf("Could not open log file %s: %s", logPath, err))
		}
		log.SetOutput(lf)
	}

	datadir := os.Getenv("ADMINAPI_TEST_DATADIR")
	if datadir == "" {
		datadir = "./testdata"
	}
	cfgFilePath := os.Getenv("ADMINAPI_TEST_CONFIGFILE")

	if cfgFilePath == "" {
		cfgFilePath = datadir + "/config.toml"
	}

	cfgFile, err := ioutil.ReadFile(cfgFilePath)
	if err != nil {
		log.Fatalf("Got error opening config file: %s", err)
	}

	//	is.randFilePath = datadir + "/10mbfile.bin"
	//	if _, err = os.Stat(is.randFilePath); os.IsNotExist(err) {
	//		f, err := os.Create(is.randFilePath)
	//		if err != nil {
	//			log.Fatalf("Cannot create file %s: %s", is.randFilePath, err)
	//		}
	//		r := rand.New(rand.NewSource(1))
	//		defer f.Close()
	//		_, err = io.CopyN(f, r, 10*1024*1024)
	//	}

	cfg := &IntegrationConfig{}
	_, err = toml.Decode(string(cfgFile), cfg)
	if err != nil {
		log.Fatalf("cannot parse config file at location '%s' : %s", cfgFile, err)
	}
	is.aa, err = NewAdminAPI(cfg.AdminAPI)
	if err != nil {
		log.Fatalf("Error initializing AdminAPI: %s", err)
	}

}

func (is *IntegrationsSuite) TearDownSuite() {
	if is.lf != nil {
		is.lf.Close()
	}
}

func (is *IntegrationsSuite) Test01Usage() {
	usage, err := is.aa.GetUsage(context.Background(), nil)
	is.NoError(err, "Got error running GetUsage")
	log.Printf("usage: %#v", usage)
}

func (is *IntegrationsSuite) Test02Metadata() {
	users, err := is.aa.MListUsers(context.Background())
	is.NoError(err, "Got error running MListUsers()")
	log.Printf("users: %#v", users)
}

func (is *IntegrationsSuite) Test03User() {
	ur := new(UserCreateRequest)
	ur.UID = "testuser"
	ur.Email = "test.user@asdf.org"
	ur.DisplayName = "Test User"
	ur.UserCaps = UserCaps{{"users", "*"}, {"metadata", "*"}, {"buckets", "read"}}
	resp, err := is.aa.UserCreate(context.Background(), ur)
	is.NoError(err, "Got error running UserCreate")
	log.Printf("%#v", resp)
	sur := new(SubUserCreateRequest)
	sur.UID = "testuser"
	sur.Access = "full"
	sur.KeyType = "s3"
	sur.SubUser = "testsubuser"
	sur.GenerateSecret = true
	nresp, err := is.aa.SubUserCreate(context.Background(), sur)
	is.NoError(err)
	log.Printf("%#v", nresp)
}

func TestIntegrations(t *testing.T) {
	suite.Run(t, new(IntegrationsSuite))
}