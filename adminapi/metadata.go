package adminapi

import (
	"context"
)

type MetaReq struct {
	Key string `url:"key"`
}

type MetaResponse struct {
	Key   string    `json:"key"`
	Mtime RadosTime `json:"mtime"`
	Ver   struct {
		Tag string `json:"tag"`
		Ver int    `json:"ver"`
	} `json:"ver"`
}

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

type MBucket struct {
	Marker        string `json:"marker"`
	Name          string `json:"name"`
	DataExtraPool string `json:"data_extra_pool"`
	BucketID      string `json:"bucket_id"`
	Pool          string `json:"pool"`
	IndexPool     string `json:"index_pool"`
	Tenant        string `json:"tenant"`
}

type Attr struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

type MUserResponse struct {
	MetaResponse
	Data UserInfoResponse `json:"data"`
}

// This is the radosgw-admin metadata list user command
// Returns a list of usernames
func (aa *AdminApi) MListUsers(ctx context.Context) ([]string, error) {
	resp := []string{}
	err := aa.get(ctx, "/metadata/user", nil, &resp)
	return resp, err
}

// This is the radosgw-admin metadata get user command
// Returns metadata about a single user
func (aa *AdminApi) MGetUser(ctx context.Context, user string) (*MUserResponse, error) {
	mr := &MetaReq{user}
	resp := &MUserResponse{}

	err := aa.get(ctx, "metadata/user", mr, resp)
	return resp, err
}

// This is the "radosgw-admin metadata list bucket" command
// Returns a list of usernames
func (aa *AdminApi) MListBuckets(ctx context.Context) ([]string, error) {
	resp := []string{}
	err := aa.get(ctx, "/metadata/bucket", nil, &resp)
	return resp, err
}

// This is the radosgw-admin metadata get bucket command
// Returns metadata about a single bucket
func (aa *AdminApi) MGetBucket(ctx context.Context, bucket string) (*MBucketResponse, error) {
	mr := &MetaReq{bucket}
	resp := &MBucketResponse{}

	err := aa.get(ctx, "metadata/bucket", mr, resp)
	return resp, err
}

// This is the "radosgw-admin metadata list bucket.instance" command
// Returns a list of usernames
func (aa *AdminApi) MListBucketInstances(ctx context.Context) ([]string, error) {
	resp := []string{}
	err := aa.get(ctx, "/metadata/bucket.instance", nil, &resp)
	return resp, err
}

// This is the radosgw-admin metadata get bucket.instance command
// Returns metadata about a single bucket.instance
func (aa *AdminApi) MGetBucketInstance(ctx context.Context, bucket string) (*MBucketInstanceResponse, error) {
	mr := &MetaReq{bucket}
	resp := &MBucketInstanceResponse{}

	err := aa.get(ctx, "metadata/bucket.instance", mr, resp)
	return resp, err
}
