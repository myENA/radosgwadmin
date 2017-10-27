[![Mozilla Public License](https://img.shields.io/badge/license-MPL-blue.svg)](https://www.mozilla.org/MPL)
[![Go Report Card](https://goreportcard.com/badge/github.com/myENA/radosgwadmin)](https://goreportcard.com/report/github.com/myENA/radosgwadmin)
[![GoDoc](https://godoc.org/github.com/myENA/radosgwadmin?status.svg)](https://godoc.org/github.com/myENA/radosgwadmin)
[![Build Status](https://travis-ci.org/myENA/radosgwadmin.svg?branch=master)](https://travis-ci.org/myENA/radosgwadmin)

Package radosgwadmin wraps http://docs.ceph.com/docs/master/radosgw/adminops

Additionally, exposes some undocumented metadata operations (methods starting with 'M').

Requires Go 1.7 or newer.  Tested with Jewel through Luminous releases of Ceph.

Example app:
```go
    package main

    import (
        "context"
        "fmt"
        "time"

        rgw "github.com/myENA/radosgwadmin"
        rcl "github.com/myENA/restclient"
    )

    func main() {

        cfg := &rgw.Config{
            ClientConfig: rcl.ClientConfig{
                ClientTimeout:   rcl.Duration(time.Second * 10),
            },
            ServerURL:       "https://my.rgw.org/",
            AdminPath:       "admin",
            AccessKeyID:     "ABC123BLAHBLAHBLAH",
            SecretAccessKey: "IMASUPERSECRETACCESSKEY",
        }

        aa, err := rgw.NewAdminAPI(cfg)
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
```

Input validation is provided by https://godoc.org/gopkg.in/go-playground/validator.v9
