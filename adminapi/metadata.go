package adminapi

import (
	"context"
)

// This is the radosgw-admin metadata list user command
// Returns a list of usernames
func (aa *AdminApi) MListUsers(ctx context.Context) ([]string, error) {
	resp := []string{}
	err := aa.Get(ctx, "/metadata/user", nil, &resp)
	return resp, err
}

// This is the "radosgw-admin metadata list bucket" command
// Returns a list of usernames
func (aa *AdminApi) MListBuckets(ctx context.Context) ([]string, error) {
	resp := []string{}
	err := aa.Get(ctx, "/metadata/bucket", nil, &resp)
	return resp, err
}

// This is the "radosgw-admin metadata list bucket.instance" command
// Returns a list of usernames
func (aa *AdminApi) MListBucketInstances(ctx context.Context) ([]string, error) {
	resp := []string{}
	err := aa.Get(ctx, "/metadata/bucket.instance", nil, &resp)
	return resp, err
}

// This is the "radosgw-admin metadata list bucket.instance" command
// Returns a list of usernames
func (aa *AdminApi) MGetBucketInstance(ctx context.Context, key string) ([]string, error) {
	resp := []string{}
	err := aa.Get(ctx, "/metadata/bucket.instance", nil, &resp)
	return resp, err
}
