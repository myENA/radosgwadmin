package adminapi

// This is the radosgw-admin metadata list user command
// Returns a list of usernames
func (aa *AdminApi) MListUsers() ([]string, error) {
	resp := []string{}
	err := aa.Get("/metadata/user", nil, &resp)
	return resp, err
}

// This is the "radosgw-admin metadata list bucket" command
// Returns a list of usernames
func (aa *AdminApi) MListBuckets() ([]string, error) {
	resp := []string{}
	err := aa.Get("/metadata/bucket", nil, &resp)
	return resp, err
}

// This is the "radosgw-admin metadata list bucket.instance" command
// Returns a list of usernames
func (aa *AdminApi) MListBucketInstances() ([]string, error) {
	resp := []string{}
	err := aa.Get("/metadata/bucket.instance", nil, &resp)
	return resp, err
}

// This is the "radosgw-admin metadata list bucket.instance" command
// Returns a list of usernames
func (aa *AdminApi) MGetBucketInstance(key string) ([]string, error) {
	resp := []string{}
	err := aa.Get("/metadata/bucket.instance", nil, &resp)
	return resp, err
}
