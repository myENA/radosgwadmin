/*
Package radosgwadmin wraps http://docs.ceph.com/docs/master/radosgw/adminops

Additionally, exposes some undocumented metadata operations.

Example app:
	cfg := &radosgwadmin.Config{
		ClientTimeout:   time.ParseDuration("30s"),
		ServerURL:       "https://my.rgw.org/",
		AdminPath:       "admin",
		AccessKeyID:     "ABC123BLAHBLAHBLAH",
		SecretAccessKey: "IMASUPERSECRETACCESSKEY",
	}

	aa := radosgwadmin.NewAdminAPI(cfg)
	users, err := aa.MListUsers(context.Background())

Input validation is provided by https://godoc.org/gopkg.in/go-playground/validator.v9,
so some errors returned will be of type validator.ValidationErrors.

*/
package radosgwadmin
