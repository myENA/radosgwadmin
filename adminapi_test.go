package radosgwadmin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
)

type ModelsSuite struct {
	suite.Suite
	dbags map[string][]byte
	vs    []interface{}
}

func (ms *ModelsSuite) SetupSuite() {
	ms.dbags = make(map[string][]byte)
	ms.vs = []interface{}{
		&quotaGetRequest{},
		&QuotaSetRequest{},
		&BucketRequest{},
		&UserCreateRequest{},
		&UserModifyRequest{},
		&SubUserCreateRequest{},
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
	resp := &BucketResponse{}
	err := json.Unmarshal(bucketjson, resp)
	ms.NoError(err, "Error unmarshaling bucket json")
	fmt.Printf("bucket response:\n%#v\n", resp)
	fmt.Printf("mktime: %s\n", time.Time(resp.Mtime).String())
}

func (ms *ModelsSuite) Test04Metadata() {
	bucketjson := ms.dbags["mbucket"]
	bresp := &MBucketResponse{}
	err := json.Unmarshal(bucketjson, bresp)
	ms.NoError(err, "Error unmarshaling mbucket json")
	fmt.Printf("mbucket response:\n%#v\n", bresp)
	userjson := ms.dbags["muser"]
	uresp := &MUserResponse{}
	err = json.Unmarshal(userjson, uresp)
	ms.NoError(err, "Error unmarshaling muser json")
	fmt.Printf("muser response:\n%#v\n", uresp)
	bucketinstjson := ms.dbags["mbucketinstance"]
	biresp := &MBucketInstanceResponse{}
	err = json.Unmarshal(bucketinstjson, biresp)
	ms.NoError(err, "Error unmarshaling mbucketinst json")
	fmt.Printf("mbucket response:\n%#v\n", biresp)
}

func (ms *ModelsSuite) Test05Quotas() {
	quotasjson := ms.dbags["quotas"]
	resp := &Quotas{}
	err := json.Unmarshal(quotasjson, resp)
	ms.NoError(err, "Error unmarshaling quotas json")
	ms.Equal(resp.BucketQuota.MaxObjects, int64(-1), "Value not expected")
	ms.Equal(resp.UserQuota.MaxSizeKb, int64(-1), "Value not expected")
}

func TestAdminAPI(t *testing.T) {
	suite.Run(t, new(ModelsSuite))
}
