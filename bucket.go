package radosgwadmin

import (
	"context"
)

// BucketRequest - bucket request struct
type BucketRequest struct {
	Bucket string `url:"bucket,omitempty"`
	UID    string `url:"uid,omitempty"`
	Stats  bool   `url:"stats,omitempty"`
}

// BucketResponse - bucket response type
type BucketResponse struct {
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

// BucketUsageEntry - bucket usage entry
type BucketUsageEntry struct {
	NumObjects   int `json:"num_objects"`
	SizeKb       int `json:"size_kb"`
	SizeKbActual int `json:"size_kb_actual"`
}

// BucketQuota - bucket quota metadata
type BucketQuota struct {
	Enabled    bool `json:"enabled"`
	MaxSizeKb  int  `json:"max_size_kb"`
	MaxObjects int  `json:"max_objects"`
}

// ListBuckets - return a list of all buckets
func (aa *AdminAPI) ListBuckets(ctx context.Context) ([]string, error) {
	resp := []string{}
	err := aa.get(ctx, "/bucket", nil, resp)
	return resp, err
}

// ListBuckets - return a list of all buckets
func (aa *AdminAPI) GetBucket(ctx context.Context, uid string) ([]string, error) {
	uir := &userInfoRequest{uid}
	resp := []string{}
	err := aa.get(ctx, "/bucket", uir, resp)
	return resp, err
}
