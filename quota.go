package radosgwadmin

import (
	"context"
)

type quotaGetRequest struct {
	UID       string `url:"uid" validate:"required"`
	QuotaType string `url:"quota-type,omitempty" validate:"omitempty,eq=user|eq=bucket"`
}

type QuotaSetRequest struct {
	UID            string `url:"uid" validate:"nonzero"`
	QuotaType      string `url:"quota-type" validate:"eq=user|eq=bucket"`
	MaximumObjects int    `url:"max-objects,omitempty"`
	MaximumSizeKb  int    `url:"max-size-kb,omitempty"`
	Enabled        bool   `url:"enabled"`
}

type QuotaMeta struct {
	Enabled    bool  `json:"enabled"`
	MaxSizeKb  int64 `json:max_size_kb"`
	MaxObjects int64 `json:max_objects"`
}

type Quotas struct {
	BucketQuota QuotaMeta `json:"bucket_quota"`
	UserQuota   QuotaMeta `json:"user_quota"`
}

func (aa *AdminAPI) Quotas(ctx context.Context, uid string) (*Quotas, error) {
	resp := &Quotas{}
	req := &quotaGetRequest{UID: uid}
	err := aa.get(ctx, "/user?quota", req, &resp)
	return resp, err
}

func (aa *AdminAPI) QuotaBucket(ctx context.Context, uid string) (*QuotaMeta, error) {
	resp := &QuotaMeta{}
	req := &quotaGetRequest{UID: uid, QuotaType: "bucket"}
	err := aa.get(ctx, "/user?quota", req, &resp)
	return resp, err
}

func (aa *AdminAPI) QuotaUser(ctx context.Context, uid string) (*QuotaMeta, error) {
	resp := &QuotaMeta{}
	req := &quotaGetRequest{UID: uid, QuotaType: "user"}
	err := aa.get(ctx, "/user?quota", req, &resp)
	return resp, err
}

func (aa *AdminAPI) SetQuota(ctx context.Context, qsr *QuotaSetRequest) error {
	return aa.put(ctx, "/user?quota", qsr, nil, nil)
}
