package radosgwadmin

import (
	"context"
)

type quotaGetRequest struct {
	UID       string `url:"uid" validate:"required"`
	QuotaType string `url:"quota-type,omitempty" validate:"omitempty,eq=user|eq=bucket"`
}

// QuotaSetRequest - passed to a QuotaSet() call
type QuotaSetRequest struct {
	UID            string `url:"uid" validate:"required"`
	QuotaType      string `url:"quota-type" validate:"eq=user|eq=bucket"`
	MaximumObjects int    `url:"max-objects,omitempty"`
	MaximumSizeKb  int    `url:"max-size-kb,omitempty"`
	Enabled        bool   `url:"enabled"`
}

// QuotaMeta - metadata about a quota
type QuotaMeta struct {
	Enabled    bool  `json:"enabled"`
	MaxSizeKb  int64 `json:"max_size_kb"`
	MaxObjects int64 `json:"max_objects"`
}

// Quotas - return type when both bucket and user quotas are returned
type Quotas struct {
	BucketQuota QuotaMeta `json:"bucket_quota"`
	UserQuota   QuotaMeta `json:"user_quota"`
}

// Quotas - get user and bucket quota info by uid
func (aa *AdminAPI) Quotas(ctx context.Context, uid string) (*Quotas, error) {
	resp := &Quotas{}
	req := &quotaGetRequest{UID: uid}
	err := aa.Get(ctx, "/user?quota", req, &resp)
	return resp, err
}

// QuotaBucket - get bucket quota info by uid
func (aa *AdminAPI) QuotaBucket(ctx context.Context, uid string) (*QuotaMeta, error) {
	resp := &QuotaMeta{}
	req := &quotaGetRequest{UID: uid, QuotaType: "bucket"}
	err := aa.Get(ctx, "/user?quota", req, &resp)
	return resp, err
}

// QuotaUser - get user quota info by uid.
func (aa *AdminAPI) QuotaUser(ctx context.Context, uid string) (*QuotaMeta, error) {
	resp := &QuotaMeta{}
	req := &quotaGetRequest{UID: uid, QuotaType: "user"}
	err := aa.Get(ctx, "/user?quota", req, &resp)
	return resp, err
}

// QuotaSet - Set a quota
func (aa *AdminAPI) QuotaSet(ctx context.Context, qsr *QuotaSetRequest) error {
	return aa.Put(ctx, "/user?quota", qsr, nil, nil)
}
