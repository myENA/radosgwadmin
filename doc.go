/*
Package radosgwadmin wraps http://docs.ceph.com/docs/master/radosgw/adminops

Additionally, exposes some undocumented metadata operations.

Example app:
	package main

	import (
		"context"
		"fmt"
		"time"

		"bitbucket.ena.net/go/radosgwadmin"
	)

	func main() {

		to, _ := time.ParseDuration("30s")
		cfg := &radosgwadmin.Config{
			ClientTimeout:   radosgwadmin.Duration{to}, // for burntsushi toml - sorry
			ServerURL:       "https://my.rgw.org/",
			AdminPath:       "admin",
			AccessKeyID:     "ABC123BLAHBLAHBLAH",
			SecretAccessKey: "IMASUPERSECRETACCESSKEY",
		}

		aa, err := radosgwadmin.NewAdminAPI(cfg)
		if err != nil {
			// do something, bail out.
		}
		users, err := aa.MListUsers(context.Background())
		if err != nil {
			// handle error
			return
		}
		fmt.Println(users)
	}

Input validation is provided by https://godoc.org/gopkg.in/go-playground/validator.v9,
so some errors returned will be of type validator.ValidationErrors.

The config struct would be a good candidate to be parsed by
https://godoc.org/github.com/BurntSushi/toml

*/
package radosgwadmin
