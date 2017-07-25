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
	"github.com/davecgh/go-spew/spew"
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

func (is *IntegrationsSuite) Test03UserCreate() {
	ur := new(UserCreateRequest)
	ur.UID = "testuser"
	ur.Email = "test.user@asdf.org"
	ur.DisplayName = "Test User"
	ur.UserCaps = UserCaps{{"users", "*"}, {"metadata", "*"}, {"buckets", "read"}}

	resp, err := is.aa.UserCreate(context.Background(), ur)
	is.NoError(err, "Got error running UserCreate")
	log.Printf("%#v", resp)
	sur := new(SubUserCreateModifyRequest)
	sur.UID = "testuser"
	sur.Access = "full"
	sur.KeyType = "s3"
	sur.SubUser = "testsubuser"
	sur.GenerateSecret = true
	nresp, err := is.aa.SubUserCreate(context.Background(), sur)
	is.NoError(err)
	log.Printf("%#v", nresp)
}

func (is *IntegrationsSuite) Test04Quota() {
	qsr := new(QuotaSetRequest)
	qsr.Enabled = true
	qsr.MaximumObjects = -1 // unlimited
	qsr.MaximumSizeKb = 8192
	qsr.QuotaType = "user"
	qsr.UID = "testuser"
	err := is.aa.QuotaSet(context.Background(), qsr)
	is.NoError(err, "Got error running SetQuota")
	// read it back
	qresp, err := is.aa.QuotaUser(context.Background(), "testuser")
	log.Printf("%#v", qresp)
	is.NoError(err, "Got error fetching user quota")
	is.True(qresp.Enabled == true, "quota not enabled")
	is.Equal(qresp.MaxObjects, int64(-1), "MaxObjects not -1")
	is.Equal(qresp.MaxSizeKb, int64(8192), "MaxSizeKb not 8192")
}

func (is *IntegrationsSuite) Test05Bucket() {
	bucketnames, err := is.aa.BucketList(context.Background(), "", "")
	is.NoError(err, "Got error fetching bucket names")
	fmt.Printf("bucket names: %#v\n", bucketnames)
	bucketstats, err := is.aa.BucketStats(context.Background(), "", "")
	is.NoError(err, "got error fetching bucket stats")

	log.Print(spew.Sdump(bucketstats))

	// TODO - make code that creates a bucket and does stuff to test
	// bucket index code. -- for now, do one I know already exists

	bireq := &BucketIndexRequest{}
	bireq.Bucket = "muhbucket"
	bireq.CheckObjects = true
	bireq.Fix = true
	bucketindresp, err := is.aa.BucketIndex(context.Background(), bireq)
	is.NoError(err, "Got error from BucketIndex()")
	log.Printf(spew.Sdump(bucketindresp))

}

func (is *IntegrationsSuite) Test05RmUser() {
	err := is.aa.UserRm(context.Background(), "testuser", true)
	is.NoError(err, "got error removing user")
	users, err := is.aa.MListUsers(context.Background())
	is.NoError(err, "got error listing users")
	found := false
	for _, user := range users {
		if user == "testuser" {
			found = true
			break
		}
	}
	is.False(found, "user not successfully deleted")
}

func TestIntegrations(t *testing.T) {
	suite.Run(t, new(IntegrationsSuite))
}
