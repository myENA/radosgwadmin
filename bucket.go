package radosgwadmin

import (
	"context"
	"encoding/json"
	"io"
)

// bucketRequest - bucket request struct
type bucketRequest struct {
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

type bucketRmRequest struct {
	Bucket       string `url:"bucket" validation:"required"`
	PurgeObjects bool   `url:"purge-objects"`
}

type bucketLinkRequest struct {
	Bucket   string `url:"bucket" validation:"required"`
	BucketID string `url:"bucket-id" validation:"required"`
	UID      string `url:"uid" validation:"required"`
}

type bucketUnlinkRequest struct {
	Bucket string `url:"bucket" validation:"required"`
	UID    string `url:"uid" validation:"required"`
}

type bucketObjectRmRequest struct {
	Bucket string `url:"bucket" validation:"required"`
	Object string `url:"object" validation:"required"`
}

type bucketPolicyRequest struct {
	Bucket string `url:"bucket" validation:"required"`
	Object string `url:"object,omitempty"`
}

// BucketIndexResponse - bucket index response struct
type BucketIndexResponse struct {
	NewObjects []string `json:"new_objects"`
	Headers    struct {
		ExistingHeader struct {
			Usage BucketUsage `json:"usage"`
		} `json:"existing_header,omitempty"`
		CalculatedHeader struct {
			Usage BucketUsage `json:"calculated_header,omitempty"`
		}
	} `json:"headers"`
}

// Decode - Implements the restapi.CustomDecoder interface
func (bir *BucketIndexResponse) Decode(data io.Reader) error {
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

// BucketUsage - Bucket usage entries
type BucketUsage struct {
	RGWNone      *BucketUsageEntry `json:"rgw.none,omitempty"`
	RGWMain      *BucketUsageEntry `json:"rgw.main,omitempty"`
	RGWShadow    *BucketUsageEntry `json:"rgw.shadow,omitempty"`
	RGWMultiMeta *BucketUsageEntry `json:"rgw.multimeta,omitempty"`
}

// BucketUsageEntry - entry for each bucket usage bit.
type BucketUsageEntry struct {
	SizeKb       uint64 `json:"size_kb"`
	SizeKbActual uint64 `json:"size_kb_actual"`
	NumObjects   uint64 `json:"num_objects"`
}

// BucketStatsResponse - bucket stats response type
type BucketStatsResponse struct {
	Bucket      string       `json:"bucket"`
	Pool        string       `json:"pool"`
	IndexPool   string       `json:"index_pool"`
	ID          string       `json:"id"`
	Marker      string       `json:"marker"`
	Owner       string       `json:"owner"`
	Ver         string       `json:"ver"`
	MasterVer   string       `json:"master_ver"`
	Mtime       RadosTime    `json:"mtime"`
	MaxMarker   string       `json:"max_marker"`
	Usage       BucketUsage  `json:"usage"`
	BucketQuota *BucketQuota `json:"bucket_quota"`
}

// BucketQuota - bucket quota metadata
type BucketQuota struct {
	Enabled    bool  `json:"enabled"`
	MaxSizeKb  int64 `json:"max_size_kb"`
	MaxObjects int64 `json:"max_objects"`
}

// BucketPolicyResponse - response from a bucket policy call
type BucketPolicyResponse struct {
	Owner struct {
		DisplayName string `json:"display_name"`
		ID          string `json:"id"`
	} `json:"owner"`
	ACL struct {
		ACLGroupMap []struct {
			ACL   int `json:"acl"`
			Group int `json:"group"`
		} `json:"acl_group_map"`
		ACLUserMap []struct {
			ACL  int    `json:"acl"`
			User string `json:"user"`
		} `json:"acl_user_map"`

		GrantMap []struct {
			ID    string `json:"id"`
			Grant struct {
				Name       string `json:"name"`
				Permission struct {
					Flags int `json:"flags"`
				} `json:"permission"`
				Type struct {
					Type int `json:"type"`
				} `json:"type"`
				Email string `json:"email"`
				ID    string `json:"id"`
				Group int    `json:"group"`
			} `json:"grant"`
		} `json:"grant_map"`
	}
}

// BucketList -
//
// return a list of all bucket names, optionally filtered by
// uid
func (aa *AdminAPI) BucketList(ctx context.Context, uid string) ([]string, error) {
	breq := &bucketRequest{
		UID:    uid,
		Bucket: "",
		Stats:  false,
	}
	resp := []string{}
	err := aa.Get(ctx, "/bucket", breq, &resp)
	return resp, err
}

// BucketStats -
//
// return a list of all bucket stats, optionally filtered by
// uid and bucket name
func (aa *AdminAPI) BucketStats(ctx context.Context, uid string, bucket string) ([]BucketStatsResponse, error) {
	resp := []BucketStatsResponse{}
	breq := &bucketRequest{
		Stats: true,
	}
	if bucket != "" {
		breq.Bucket = bucket
		respB := BucketStatsResponse{}
		err := aa.Get(ctx, "/bucket", breq, &respB)
		return append(resp, respB), err
	}

	breq.UID = uid
	err := aa.Get(ctx, "/bucket", breq, &resp)
	return resp, err
}

// BucketIndex - Bucket index operations.  Bucket name required.
func (aa *AdminAPI) BucketIndex(ctx context.Context, bireq *BucketIndexRequest) (*BucketIndexResponse, error) {
	resp := &BucketIndexResponse{}
	err := aa.Get(ctx, "/bucket?index", bireq, resp)
	return resp, err
}

// BucketRm - remove a bucket.  bucket must be non-empty string.
func (aa *AdminAPI) BucketRm(ctx context.Context, bucket string, purge bool) error {
	req := &bucketRmRequest{Bucket: bucket, PurgeObjects: purge}
	return aa.Delete(ctx, "/bucket", req, nil)
}

// BucketUnlink - unlink a bucket from a user.  All parameters required.
func (aa *AdminAPI) BucketUnlink(ctx context.Context, bucket string, uid string) error {
	req := &bucketUnlinkRequest{Bucket: bucket, UID: uid}
	return aa.Post(ctx, "/bucket", req, nil, nil)
}

// BucketLink - link a bucket to a user, removing any previous links.  All
// parameters required.
func (aa *AdminAPI) BucketLink(ctx context.Context, bucket, bucketID, uid string) error {
	req := &bucketLinkRequest{Bucket: bucket, BucketID: bucketID, UID: uid}
	return aa.Put(ctx, "/bucket", req, nil, nil)
}

// BucketObjectRm - remove a bucket.  bucket must be non-empty string.
func (aa *AdminAPI) BucketObjectRm(ctx context.Context, bucket, object string) error {
	req := &bucketObjectRmRequest{Bucket: bucket, Object: object}
	return aa.Delete(ctx, "/bucket?object", req, nil)
}

// BucketPolicy - get a bucket policy.  bucket required, object is optional.
func (aa *AdminAPI) BucketPolicy(ctx context.Context, bucket, object string) (*BucketPolicyResponse, error) {
	req := bucketPolicyRequest{Bucket: bucket, Object: object}
	resp := &BucketPolicyResponse{}
	err := aa.Get(ctx, "/bucket?policy", req, resp)
	return resp, err
}
