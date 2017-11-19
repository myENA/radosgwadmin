package radosgwadmin

import (
	"context"
)

type metaReq struct {
	Key string `url:"key"`
}

// MetaResponse - base response type of all metadata calls.
type MetaResponse struct {
	Key   string    `json:"key"`
	Mtime RadosTime `json:"mtime"`
	Ver   struct {
		Tag string `json:"tag"`
		Ver int    `json:"ver"`
	} `json:"ver"`
}

// MBucketInstanceResponse - response from bucket.instance get
type MBucketInstanceResponse struct {
	MetaResponse
	Data struct {
		BucketInfo struct {
			Bucket            MBucket     `json:"bucket"`
			NumShards         int         `json:"num_shards"`
			PlacementRule     string      `json:"placement_rule"`
			SwiftVerLocation  string      `json:"swift_ver_location"`
			Flags             int         `json:"flags"`
			HasWebsite        string      `json:"has_website"` // bad bool
			Quota             BucketQuota `json:"quota"`
			SwiftVersioning   string      `json:"swift_versioning"`
			Owner             string      `json:"owner"`
			RequesterPays     string      `json:"requester_pays"`
			IndexType         int         `json:"index_type"`
			BiShardHashType   int         `json:"bi_shard_hash_type"`
			HasInstanceObject string      `json:"has_instance_obj"`
			CreationTime      string      `json:"creation_time"`
			ZoneGroup         string      `json:"zonegroup"`
		} `json:"bucket_info"`
		Attrs []Attr `json:"attrs"`
	} `json:"data"`
}

// MBucketResponse - response from metadata bucket get
type MBucketResponse struct {
	MetaResponse
	Data struct {
		Bucket        MBucket `json:"bucket"`
		HasBucketInfo string  `json:"has_bucket_info"`
		Linked        string  `json:"linked"`        // bad bool
		CreationTime  string  `json:"creation_time"` // bad float
		Owner         string  `json:"owner"`
	} `json:"data"`
}

// MBucket - bucket information
type MBucket struct {
	Marker        string `json:"marker"`
	Name          string `json:"name"`
	DataExtraPool string `json:"data_extra_pool"`
	BucketID      string `json:"bucket_id"`
	Pool          string `json:"pool"`
	IndexPool     string `json:"index_pool"`
	Tenant        string `json:"tenant"`
}

// Attr - kv data that acompanies some metadata responses.
type Attr struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

// MUserResponse - user metadata response
type MUserResponse struct {
	MetaResponse
	Data UserInfoResponse `json:"data"`
}

// MListUsers - This is the radosgw-admin metadata list user command
// Returns a list of usernames
func (aa *AdminAPI) MListUsers(ctx context.Context) ([]string, error) {
	resp := []string{}
	err := aa.Get(ctx, "/metadata/user", nil, &resp)
	return resp, err
}

// MGetUser - This is the radosgw-admin metadata get user command
// Returns metadata about a single user
func (aa *AdminAPI) MGetUser(ctx context.Context, user string) (*MUserResponse, error) {
	mr := &metaReq{user}
	resp := &MUserResponse{}

	err := aa.Get(ctx, "metadata/user", mr, resp)
	return resp, err
}

// MListBuckets - This is the "radosgw-admin metadata list bucket" command
// Returns a list of usernames
func (aa *AdminAPI) MListBuckets(ctx context.Context) ([]string, error) {
	resp := []string{}
	err := aa.Get(ctx, "/metadata/bucket", nil, &resp)
	return resp, err
}

// MGetBucket - This is the radosgw-admin metadata get bucket command
// Returns metadata about a single bucket
func (aa *AdminAPI) MGetBucket(ctx context.Context, bucket string) (*MBucketResponse, error) {
	mr := &metaReq{bucket}
	resp := &MBucketResponse{}

	err := aa.Get(ctx, "metadata/bucket", mr, resp)
	return resp, err
}

// MListBucketInstances - This is the "radosgw-admin metadata list bucket.instance" command
// Returns a list of usernames
func (aa *AdminAPI) MListBucketInstances(ctx context.Context) ([]string, error) {
	resp := []string{}
	err := aa.Get(ctx, "/metadata/bucket.instance", nil, &resp)
	return resp, err
}

// MGetBucketInstance - This is the radosgw-admin metadata get bucket.instance command
// Returns metadata about a single bucket.instance
func (aa *AdminAPI) MGetBucketInstance(ctx context.Context, bucket string) (*MBucketInstanceResponse, error) {
	mr := &metaReq{bucket}
	resp := &MBucketInstanceResponse{}

	err := aa.Get(ctx, "metadata/bucket.instance", mr, resp)
	return resp, err
}
