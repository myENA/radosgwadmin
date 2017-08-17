// +build integration

package radosgwadmin

import (
	"context"
	"io/ioutil"
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
	i            *Integration
}

type IntegrationConfig struct {
	Integration *Integration
	RGW         *Config
}

type Integration struct {
	TestUID         string
	TestEmail       string
	TestDisplayName string
	TestSubUser     string
	TesTSubUser2    string
	TestBucket      string // needs to exist already
}

func (is *IntegrationsSuite) SetupSuite() {

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
		is.T().Logf("Got error opening config file: %s", err)
		os.Exit(1)
	}

	cfg := &IntegrationConfig{}
	_, err = toml.Decode(string(cfgFile), cfg)

	if err != nil {
		is.T().Logf("cannot parse config file at location '%s' : %s", cfgFile, err)
		os.Exit(1)
	}
	is.aa, err = NewAdminAPI(cfg.RGW)
	is.i = cfg.Integration
	if err != nil {
		is.T().Logf("Error initializing AdminAPI: %s", err)
		os.Exit(1)
	}

}

func (is *IntegrationsSuite) TearDownSuite() {
	if is.lf != nil {
		is.lf.Close()
	}
}

func (is *IntegrationsSuite) Test01Usage() {
	usage, err := is.aa.Usage(context.Background(), nil)
	is.NoError(err, "Got error getting Usage")
	is.T().Logf("usage: %#v", usage)
	err = is.aa.UsageTrim(context.Background(), &TrimUsageRequest{UID: is.i.TestUID})
	is.NoError(err, "Got error trimming usage")
}

func (is *IntegrationsSuite) Test02Metadata() {
	users, err := is.aa.MListUsers(context.Background())
	is.NoError(err, "Got error running MListUsers()")
	is.T().Logf("users: %#v", users)
}

func (is *IntegrationsSuite) Test03UserCreate() {
	ur := new(UserCreateRequest)
	ur.UID = is.i.TestUID
	ur.Email = is.i.TestEmail
	ur.DisplayName = is.i.TestDisplayName
	ur.UserCaps = []UserCap{{"users", "*"}, {"metadata", "*"}, {"buckets", "read"}}

	resp, err := is.aa.UserCreate(context.Background(), ur)
	is.NoError(err, "Got error running UserCreate")
	is.T().Logf("%#v", resp)
	sur := new(SubUserCreateModifyRequest)
	sur.UID = is.i.TestUID
	sur.Access = "full"
	sur.KeyType = "s3"
	sur.SubUser = is.i.TestSubUser
	sur.GenerateSecret = true
	nresp, err := is.aa.SubUserCreate(context.Background(), sur)
	is.NoError(err)
	is.T().Logf("%#v", nresp)
	sur.SubUser = is.i.TesTSubUser2
	sur.Access = "read"
	nresp, err = is.aa.SubUserCreate(context.Background(), sur)
	is.NoError(err)
	is.T().Logf("%#v", nresp)
}

func (is *IntegrationsSuite) Test04Quota() {
	qsr := new(QuotaSetRequest)
	qsr.Enabled = true
	qsr.MaximumObjects = -1 // unlimited
	qsr.MaximumSizeKb = 8192
	qsr.QuotaType = "user"
	qsr.UID = is.i.TestUID
	err := is.aa.QuotaSet(context.Background(), qsr)
	is.NoError(err, "Got error running SetQuota")
	// read it back
	qresp, err := is.aa.QuotaUser(context.Background(), is.i.TestUID)
	is.T().Logf("%#v", qresp)
	is.NoError(err, "Got error fetching user quota")
	is.True(qresp.Enabled == true, "quota not enabled")
	is.Equal(qresp.MaxObjects, int64(-1), "MaxObjects not -1")
	is.Equal(qresp.MaxSizeKb, int64(8192), "MaxSizeKb not 8192")
}

func (is *IntegrationsSuite) Test05Bucket() {
	bucketnames, err := is.aa.BucketList(context.Background(), "")
	is.NoError(err, "Got error fetching bucket names")
	is.T().Logf("bucket names: %#v\n", bucketnames)
	bucketstats, err := is.aa.BucketStats(context.Background(), "", "")
	is.NoError(err, "got error fetching bucket stats")

	is.T().Log(spew.Sdump(bucketstats))

	// TODO - make code that creates a bucket and does stuff to test
	// bucket index code. -- for now, do one I know already exists

	bireq := &BucketIndexRequest{}
	bireq.Bucket = is.i.TestBucket
	bireq.CheckObjects = true
	bireq.Fix = true
	bucketindresp, err := is.aa.BucketIndex(context.Background(), bireq)
	is.NoError(err, "Got error from BucketIndex()")
	is.T().Logf(spew.Sdump(bucketindresp))

}

func (is *IntegrationsSuite) Test06Caps() {
	ucr := &UserCapsRequest{
		UID:      is.i.TestUID,
		UserCaps: []UserCap{{"usage", "read"}},
	}
	newcaps, err := is.aa.CapsAdd(context.Background(), ucr)
	is.NoError(err, "Unexpected error adding capabilities")
	is.Len(newcaps, 4, "unexpected len")
	found := false
	for _, cap := range newcaps {
		if cap.String() == "usage=read" {
			found = true
			break
		}
	}
	is.True(found, "could not find the permission we just added")
	ucr.UserCaps = []UserCap{{"usage", "write"}}

	found = false
	newcaps, err = is.aa.CapsAdd(context.Background(), ucr)
	is.NoError(err, "unexpected error")
	for _, cap := range newcaps {
		if cap.String() == "usage=*" {
			found = true
			break
		}
	}
	is.True(found, "Permissions are not additive like we thought")

	ucr.UserCaps = []UserCap{{"usage", "write"}, {"metadata", "write"}}

	newcaps, err = is.aa.CapsRm(context.Background(), ucr)
	is.NoError(err, "unexpected error")
	goodct := 0
	for _, caps := range newcaps {
		switch caps.String() {
		case "metadata=read", "usage=read":
			goodct++
		}
	}
	is.Equal(goodct, 2, "not expected removal of perms")

}

func (is *IntegrationsSuite) Test07RmUser() {
	err := is.aa.UserRm(context.Background(), is.i.TestUID, true)
	is.NoError(err, "got error removing user")
	users, err := is.aa.MListUsers(context.Background())
	is.NoError(err, "got error listing users")
	found := false
	for _, user := range users {
		if user == is.i.TestUID {
			found = true
			break
		}
	}
	is.False(found, "user not successfully deleted")
}

func TestIntegrations(t *testing.T) {
	suite.Run(t, new(IntegrationsSuite))
}
