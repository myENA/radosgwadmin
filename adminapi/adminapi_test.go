package adminapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ModelsSuite struct {
	suite.Suite
	dbags map[string][]byte
}

func (ms *ModelsSuite) SetupSuite() {
	ms.dbags = make(map[string][]byte)
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

func (suite *ModelsSuite) Test01Usage() {
	usagejson := suite.dbags["usage"]
	resp := &UsageResponse{}
	err := json.Unmarshal(usagejson, resp)
	suite.NoError(err, "Error unmarshaling json")
	suite.Len(resp.Entries, 1, "Expected number of entries not found")
	suite.Len(resp.Entries[0].Buckets, 2, "Expected number of Buckets not found")
	t, err := time.Parse(RadosTimeFormat, "2017-03-16 04:00:00.000000Z")
	suite.NoError(err, "Error received when no error expected")
	suite.Equal(time.Time(resp.Entries[0].Buckets[0].Time), t, "Time formats don't match")
	suite.Len(resp.Summary, 1, "Expected summary size not found")
	suite.Len(resp.Summary[0].Categories, 6, "Expected number of categories in the summary not found")

}

func (suite *ModelsSuite) Test02Bucket() {
	bucketjson := suite.dbags["bucket"]
	resp := &BucketResponse{}
	err := json.Unmarshal(bucketjson, resp)
	suite.NoError(err, "Error unmarshaling bucket json")
	fmt.Printf("bucket response:\n%#v\n", resp)
	fmt.Printf("mktime: %s\n", time.Time(resp.Mtime).String())
}

func (suite *ModelsSuite) Test03Metadata() {
	bucketjson := suite.dbags["mbucket"]
	bresp := &MBucketResponse{}
	err := json.Unmarshal(bucketjson, bresp)
	suite.NoError(err, "Error unmarshaling mbucket json")
	fmt.Printf("mbucket response:\n%#v\n", bresp)
	userjson := suite.dbags["muser"]
	uresp := &MUserResponse{}
	err = json.Unmarshal(userjson, uresp)
	suite.NoError(err, "Error unmarshaling muser json")
	fmt.Printf("muser response:\n%#v\n", uresp)
	bucketinstjson := suite.dbags["mbucketinstance"]
	biresp := &MBucketInstanceResponse{}
	err = json.Unmarshal(bucketinstjson, biresp)
	suite.NoError(err, "Error unmarshaling mbucketinst json")
	fmt.Printf("mbucket response:\n%#v\n", biresp)
}

func TestAdminApi(t *testing.T) {
	suite.Run(t, new(ModelsSuite))
}
