// +build integration

package radosgwadmin

import (
	"context"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/BurntSushi/toml"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/suite"
)

type IntegrationsSuite struct {
	suite.Suite
	aa           *AdminAPI
	randFilePath string
	lf           *os.File
	ic           *IntegrationConfig
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

	is.ic = cfg
	is.aa, err = NewAdminAPI(cfg.RGW)
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
	err = is.aa.UsageTrim(context.Background(), &TrimUsageRequest{UID: is.ic.Integration.TestUID})
	is.NoError(err, "Got error trimming usage")
}

func (is *IntegrationsSuite) Test02Metadata() {
	users, err := is.aa.MListUsers(context.Background())
	is.NoError(err, "Got error running MListUsers()")
	is.T().Logf("users: %#v", users)
}

func (is *IntegrationsSuite) Test03UserCreate() {
	ur := new(UserCreateRequest)
	ur.UID = is.ic.Integration.TestUID
	ur.Email = is.ic.Integration.TestEmail
	ur.DisplayName = is.ic.Integration.TestDisplayName
	ur.UserCaps = []UserCap{{"users", "*"}, {"metadata", "*"}, {"buckets", "read"}}

	resp, err := is.aa.UserCreate(context.Background(), ur)
	is.NoError(err, "Got error running UserCreate")
	is.T().Logf("%#v", resp)
	sur := new(SubUserCreateModifyRequest)
	sur.UID = is.ic.Integration.TestUID
	sur.Access = "full"
	sur.KeyType = "s3"
	sur.SubUser = is.ic.Integration.TestSubUser
	sur.GenerateSecret = true
	nresp, err := is.aa.SubUserCreate(context.Background(), sur)
	is.NoError(err)
	is.T().Logf("%#v", nresp)
	sur.SubUser = is.ic.Integration.TesTSubUser2
	sur.Access = "read"
	nresp, err = is.aa.SubUserCreate(context.Background(), sur)
	is.NoError(err)
	is.T().Logf("%#v", nresp)
}

func (is *IntegrationsSuite) Test04Quota() {
	qsr := new(QuotaSetRequest)
	qsr.Enabled = true
	qsr.MaximumObjects = -1 // unlimited
	qsr.MaximumSizeKb = 61440
	qsr.QuotaType = "user"
	qsr.UID = is.ic.Integration.TestUID
	err := is.aa.QuotaSet(context.Background(), qsr)
	is.NoError(err, "Got error running SetQuota")
	// read it back
	qresp, err := is.aa.QuotaUser(context.Background(), is.ic.Integration.TestUID)
	is.T().Logf("%#v", qresp)
	is.NoError(err, "Got error fetching user quota")
	is.True(qresp.Enabled == true, "quota not enabled")
	is.Equal(qresp.MaxObjects, int64(-1), "MaxObjects not -1")
	is.Equal(qresp.MaxSizeKb, int64(61440), "MaxSizeKb not 61440")
}

func (is *IntegrationsSuite) Test05Bucket() {
	// Write 50mb to it.
	is.writeRandomFile("bigrandomfile.bin", 1024*10)
	is.aa.BucketIndex(context.Background(), &BucketIndexRequest{
		Bucket:       is.ic.Integration.TestBucket,
		CheckObjects: true,
		Fix:          true,
	})

	bucketnames, err := is.aa.BucketList(context.Background(), "")
	is.NoError(err, "Got error fetching bucket names")
	is.T().Logf("bucket names: %#v\n", bucketnames)
	// bucketstats with no filters
	bucketstats, err := is.aa.BucketStats(context.Background(), "", "")
	is.NoError(err, "got error fetching bucket stats")
	is.T().Log(spew.Sdump(bucketstats))

	// bucketstats with bucket filter
	bucketstatsf, err := is.aa.BucketStats(context.Background(), "", is.ic.Integration.TestBucket)
	is.NoError(err, "got error fetching bucket stats filtered by bucket")
	is.T().Log(spew.Sdump(bucketstatsf))

	// TODO: - make code that creates a bucket and does stuff to test
	// bucket index code. -- for now, do one I know already exists

	bireq := &BucketIndexRequest{}
	bireq.Bucket = is.ic.Integration.TestBucket
	bireq.CheckObjects = true
	bireq.Fix = true
	bucketindresp, err := is.aa.BucketIndex(context.Background(), bireq)
	is.NoError(err, "Got error from BucketIndex()")
	is.T().Logf(spew.Sdump(bucketindresp))

	// see if we can now get stats.

	ui, err := is.aa.UserInfo(context.Background(), is.ic.Integration.TestUID, true)

	is.NoError(err)
	is.NotNil(ui.Stats)

	is.T().Log(spew.Sdump(ui))
}

func (is *IntegrationsSuite) Test06Caps() {
	ucr := &UserCapsRequest{
		UID:      is.ic.Integration.TestUID,
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

func (is *IntegrationsSuite) Test07Keys() {
	accessKey := "TESTACCESSKEY"
	secretKey := "TESTSECRETKEY"
	generateKey := false
	kc := &KeyCreateRequest{
		UID:         is.ic.Integration.TestUID,
		AccessKey:   accessKey,
		SecretKey:   secretKey,
		GenerateKey: &generateKey,
	}

	userKeys, err := is.aa.KeyCreate(context.Background(), kc)
	is.NoError(err, "Unexpected error adding key")
	found := false
	for _, key := range userKeys {
		if key.AccessKey == accessKey && key.SecretKey == secretKey {
			found = true
			break
		}
	}
	is.True(found, "could not find the key we just added")

	// now try to remove the key
	kr := &KeyRmRequest{
		AccessKey: accessKey,
	}

	err1 := is.aa.KeyRm(context.Background(), kr)
	is.NoError(err1, "Unexpected error deleting key")

	// check if the key was really deleted
	userInfo, err2 := is.aa.UserInfo(context.Background(), is.ic.Integration.TestUID, false)
	is.NoError(err2, "Unexpected error reading user info")
	found = false
	for i := range userInfo.Keys {
		if ok := userInfo.Keys[i].AccessKey == accessKey; ok {
			found = true
			break
		}
	}
	is.False(found, "still could find the key we just deleted")
}

func (is *IntegrationsSuite) Test08RmUser() {
	leave := os.Getenv("LEAVE_USER")
	if leave != "" {
		l, _ := strconv.ParseBool(leave)
		if l {
			is.T().Log("Skipping user delete")
			return
		}
	}
	err := is.aa.UserRm(context.Background(), is.ic.Integration.TestUID, true)
	is.NoError(err, "got error removing user")
	users, err := is.aa.MListUsers(context.Background())
	is.NoError(err, "got error listing users")
	found := false
	for _, user := range users {
		if user == is.ic.Integration.TestUID {
			found = true
			break
		}
	}
	is.False(found, "user not successfully deleted")
}

func (is *IntegrationsSuite) writeRandomFile(path string, sizekb int) {
	var (
		s3c   *s3.S3
		err   error
		ui    *UserInfoResponse
		creds *credentials.Credentials
	)

	// grab creds for subuser s3 user
	ui, err = is.aa.UserInfo(context.Background(), is.ic.Integration.TestUID, false)

	is.NoError(err)
	for _, k := range ui.Keys {
		if k.User == is.ic.Integration.TestUID {
			creds = credentials.NewStaticCredentials(k.AccessKey, k.SecretKey, "")
			break
		}
	}

	is.NotNil(creds, "credentials not found")
	if creds == nil {
		return
	}
	cfg := aws.NewConfig()
	cfg.Region = aws.String("us-east")
	cfg.Endpoint = aws.String(is.ic.RGW.ServerURL)
	cfg.WithS3ForcePathStyle(true)

	cfg.Credentials = creds
	sess := session.Must(session.NewSession(cfg))
	s3c = s3.New(sess)

	_, err = s3c.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(is.ic.Integration.TestBucket),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(""),
		},
	})
	if err != nil {

		// ignore non-fatal creation errors
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() != s3.ErrCodeBucketAlreadyExists &&
				awsErr.Code() != s3.ErrCodeBucketAlreadyOwnedByYou {
				is.FailNow(err.Error())
			}
		} else {
			is.FailNow(err.Error())
		}
	}

	rr := NewRandoReader(0, int64(1024*sizekb))

	uploader := s3manager.NewUploader(sess)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(is.ic.Integration.TestBucket),
		Key:    aws.String(path),
		Body:   rr,
	})
	if err != nil {
		is.Fail(err.Error())
	}

}

type RandoReader struct {
	r    *rand.Rand
	max  int64
	read int64
}

func NewRandoReader(seed int64, bytes int64) *RandoReader {
	return &RandoReader{
		r:    rand.New(rand.NewSource(seed)),
		max:  bytes,
		read: 0,
	}
}

func (rr *RandoReader) Read(p []byte) (n int, err error) {
	if rr.read >= rr.max {
		return 0, io.EOF
	} else if rr.max < (rr.read + int64(len(p))) {
		p = p[:(rr.max - rr.read)]
	}
	c, err := rr.r.Read(p)
	rr.read += int64(c)
	return c, err
}

func TestIntegrations(t *testing.T) {
	suite.Run(t, new(IntegrationsSuite))
}
