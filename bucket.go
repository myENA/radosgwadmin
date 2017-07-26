package radosgwadmin

import (
	"context"
	"encoding/json"
	"io"
)

// BucketRequest - bucket request struct
type BucketRequest struct {
	Bucket string `url:"bucket,omitempty"`
	UID    string `url:"uid,omitempty"`
	Stats  bool   `url:"stats,omitempty"`
}

// BucketIndexRequest - bucket index request struct
type BucketIndexRequest struct {
	Bucket       string `url:"bucket" validation:"required"`
	CheckObjects bool   `url:"check-objects,omitempty"`
	Fix          bool   `url:"fix,omitempty"`
}

// BucketIndexResponse - bucket index response struct
type BucketIndexResponse struct {
	NewObjects []string `json:"new_objects"`
	Headers    struct {
		ExistingHeader   *BucketUsage `json:"existing_header,omitempty"`
		CalculatedHeader *BucketUsage `json:"calculated_header,omitempty"`
	} `json:"headers"`
}

// Implements the customDecoder interface
func (bir *BucketIndexResponse) decode(data io.Reader) error {
	// rgw does some weird shit with this response.
	// has an array followed by a json object, no delimters.
	dec := json.NewDecoder(data)

	// Read the array first.
	err := dec.Decode(&bir.NewObjects)
	if err != nil {
		return err
	}

	// Now read the object.
	if dec.More() {
		err = dec.Decode(&bir.Headers)
		if err != nil {
			return err
		}
	}
	return nil
}

// BucketUsage - bucket usage collection
type BucketUsage struct {
	Usage struct {
		RGWNone      *BucketUsageEntry `json:"rgw.none,omitempty"`
		RGWMain      *BucketUsageEntry `json:"rgw.main,omitempty"`
		RGWShadow    *BucketUsageEntry `json:"rgw.shadow,omitempty"`
		RGWMultiMeta *BucketUsageEntry `json:"rgw.multimeta,omitempty"`
	} `json:"usage"`
}

// BucketUsageEntry - entry for each bucket usage bit.
type BucketUsageEntry struct {
	SizeKb       int `json:"size_kb"`
	SizeKbActual int `json:"size_kb_actual"`
	NumObjects   int `json:"num_objects"`
}

// BucketStatsResponse - bucket stats response type
type BucketStatsResponse struct {
	Bucket      string                      `json:"bucket"`
	Pool        string                      `json:"pool"`
	IndexPool   string                      `json:"index_pool"`
	ID          string                      `json:"id"`
	Marker      string                      `json:"marker"`
	Owner       string                      `json:"owner"`
	Ver         string                      `json:"ver"`
	MasterVer   string                      `json:"master_ver"`
	Mtime       RadosTime                   `json:"mtime"`
	MaxMarker   string                      `json:"max_marker"`
	Usage       map[string]BucketUsageEntry `json:"usage"`
	BucketQuota *BucketQuota                `json:"bucket_quota"`
}

// BucketQuota - bucket quota metadata
type BucketQuota struct {
	Enabled    bool `json:"enabled"`
	MaxSizeKb  int  `json:"max_size_kb"`
	MaxObjects int  `json:"max_objects"`
}

// BucketList -
//
// return a list of all bucket names, optionally filtered by
// uid and bucket name
func (aa *AdminAPI) BucketList(ctx context.Context, uid string, bucket string) ([]string, error) {
	breq := &BucketRequest{
		UID:    uid,
		Bucket: bucket,
		Stats:  false,
	}
	resp := []string{}
	err := aa.get(ctx, "/bucket", breq, &resp)
	return resp, err
}

// BucketStats -
//
// return a list of all bucket stats, optionally filtered by
// uid and bucket name
func (aa *AdminAPI) BucketStats(ctx context.Context, uid string, bucket string) ([]BucketStatsResponse, error) {
	breq := &BucketRequest{
		UID:    uid,
		Bucket: bucket,
		Stats:  true,
	}
	resp := []BucketStatsResponse{}
	err := aa.get(ctx, "/bucket", breq, &resp)
	return resp, err
}

// BucketIndex - Bucket index operations.  Bucket name required.
func (aa *AdminAPI) BucketIndex(ctx context.Context, bireq *BucketIndexRequest) (*BucketIndexResponse, error) {
	resp := &BucketIndexResponse{}
	err := aa.get(ctx, "/bucket?index", bireq, resp)
	return resp, err
}
