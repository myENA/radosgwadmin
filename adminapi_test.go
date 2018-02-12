package radosgwadmin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
)

var validate *validator.Validate

type ModelsSuite struct {
	suite.Suite
	dbags map[string][]byte
	vs    []interface{}
	aa    *AdminAPI
}

func (ms *ModelsSuite) SetupSuite() {
	ms.dbags = make(map[string][]byte)
	ms.vs = []interface{}{
		&quotaGetRequest{},
		&QuotaSetRequest{},
		&UserCreateRequest{},
		&UserCapsRequest{},
		&UserModifyRequest{},
		&SubUserCreateModifyRequest{},
		&bucketRequest{},
		&bucketRmRequest{},
		&bucketLinkRequest{},
		&bucketUnlinkRequest{},
		&BucketIndexRequest{},
		&bucketObjectRmRequest{},
	}
	datadir := os.Getenv("ADMINAPI_TEST_DATADIR")
	if datadir == "" {
		datadir = "./testdata"
	}
	files, err := ioutil.ReadDir(datadir)
	if err != nil {
		panic(fmt.Sprintf("Got error trying to open dir %s: %s", datadir, err))
	}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".json") {
			continue
		}
		path := datadir + "/" + file.Name()
		key := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		ms.dbags[key], err = ioutil.ReadFile(path)
		if err != nil {
			panic(fmt.Sprintf("Got error trying to read file %s: %s", path, err))
		}
	}

	ms.aa = &AdminAPI{}
	validate = validator.New()

}

func (ms *ModelsSuite) Test01Validators() {
	var structName string
	defer func() {
		if r := recover(); r != nil {
			ms.Fail("Paniced in Validators()", "paniced validating %s: %#v\n", structName, r)

		}
	}()
	for _, v := range ms.vs {
		structName = reflect.TypeOf(v).Elem().Name()
		err := validate.Struct(v)
		if err != nil {
			vierr, ok := err.(*validator.InvalidValidationError)
			ms.False(ok, "Error is InvalidValidationError %s", vierr)
		}
	}

	// test the UserCreateRequest validators
	ucr := &UserCreateRequest{
		UID:     "whatever",
		Email:   "asdf@asdf.org",
		KeyType: "swift",
		UserCaps: []UserCap{
			{
				Type:       "buckets",
				Permission: "read,write",
			},
			{
				Type:       "asdf",
				Permission: "invalid",
			},
		},
	}
	err := validate.Struct(ucr)
	ms.Error(err, "did not fail validation")
	_, ok := err.(validator.ValidationErrors)
	ms.True(ok, "not coerce to ValidationErrors")

}

func (ms *ModelsSuite) Test02Usage() {
	usagejson := ms.dbags["usage"]
	resp := &UsageResponse{}
	err := json.Unmarshal(usagejson, resp)
	ms.NoError(err, "Error unmarshaling json")
	ms.Len(resp.Entries, 1, "Expected number of entries not found")
	ms.Len(resp.Entries[0].Buckets, 2, "Expected number of Buckets not found")
	t, err := time.Parse(RadosTimeFormat, "2017-03-16 04:00:00.000000Z")
	ms.NoError(err, "Error received when no error expected")
	ms.Equal(time.Time(resp.Entries[0].Buckets[0].Time), t, "Time formats don't match")
	ms.Len(resp.Summary, 1, "Expected summary size not found")
	ms.Len(resp.Summary[0].Categories, 6, "Expected number of categories in the summary not found")

}

func (ms *ModelsSuite) Test03Bucket() {
	bucketjson := ms.dbags["bucket"]
	resp := &BucketStatsResponse{}
	err := json.Unmarshal(bucketjson, resp)
	ms.NoError(err, "Error unmarshaling bucket json")
	ms.T().Logf("bucket response:\n%#v\n", resp)
	ms.T().Logf("mktime: %s\n", time.Time(resp.Mtime).String())

	bucketindjson := ms.dbags["bucketindex"]
	bir := &BucketIndexResponse{}
	err = bir.Decode(bytes.NewReader(bucketindjson))
	ms.NoError(err, "Error unmarshaling bucket index json")
	ms.Equal(bir.NewObjects[0], "key.json", "first element of NewObjects not as expected")
	ms.Equal(len(bir.NewObjects), 3, "length of NewObjects not 3")
	ms.Equal(bir.Headers.ExistingHeader.Usage.RGWMain.NumObjects, uint64(9), "rgwmain num objects not as expected")
	ms.Equal(bir.Headers.ExistingHeader.Usage.RGWNone.SizeKb, uint64(5), "rgwnone num objects not as expected")

	bucketindjsonNoFix := ms.dbags["bucketindex_nofix"]
	bir = &BucketIndexResponse{}
	err = bir.Decode(bytes.NewReader(bucketindjsonNoFix))
	ms.NoError(err, "Error, could not read bucketindex_nofix")

	bucketpoljson := ms.dbags["bucketpolicy"]
	bpr := &BucketPolicyResponse{}
	err = json.Unmarshal(bucketpoljson, bpr)
	ms.NoError(err, "Error, could not read from bucket policy")

	ms.T().Log(spew.Sdump(bpr))
}

func (ms *ModelsSuite) Test04Metadata() {
	bucketjson := ms.dbags["mbucket"]
	bresp := &MBucketResponse{}
	err := json.Unmarshal(bucketjson, bresp)
	ms.NoError(err, "Error unmarshaling mbucket json")
	ms.T().Logf("mbucket response:\n%#v\n", bresp)
	userjson := ms.dbags["muser"]
	uresp := &MUserResponse{}
	err = json.Unmarshal(userjson, uresp)
	ms.NoError(err, "Error unmarshaling muser json")
	ms.T().Logf("muser response:\n%#v\n", uresp)
	bucketinstjson := ms.dbags["mbucketinstance"]
	biresp := &MBucketInstanceResponse{}
	err = json.Unmarshal(bucketinstjson, biresp)
	ms.NoError(err, "Error unmarshaling mbucketinst json")
	ms.T().Logf("mbucket response:\n%#v\n", biresp)
}

func (ms *ModelsSuite) Test05Quotas() {
	quotasjson := ms.dbags["quotas"]
	resp := &Quotas{}
	err := json.Unmarshal(quotasjson, resp)
	ms.NoError(err, "Error unmarshaling quotas json")
	ms.Equal(resp.BucketQuota.MaxObjects, int64(-1), "Value not expected")
	ms.Equal(resp.UserQuota.MaxSizeKb, int64(-1), "Value not expected")
}

func (ms *ModelsSuite) Test06User() {
	userjson := ms.dbags["user"]
	userstatjson := ms.dbags["userstat"]
	resp := &UserInfoResponse{}
	err := json.Unmarshal(userjson, resp)
	ms.NoError(err, "Error unmarshaling user json")
	ms.Nil(resp.Stats, "stats not nil as expected")
	resp = &UserInfoResponse{}
	err = json.Unmarshal(userstatjson, resp)
	ms.NoError(err, "Error unmsrshaling userstat json")
	ms.NotNil(resp.Stats, "userstat Stats is nil when it shouldn't be")

}

func TestAdminAPI(t *testing.T) {
	suite.Run(t, new(ModelsSuite))
}
