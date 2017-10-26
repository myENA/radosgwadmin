/*
Package radosgwadmin wraps http://docs.ceph.com/docs/master/radosgw/adminops

Additionally, exposes some undocumented metadata operations (methods starting with 'M').

Example app:
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

Input validation is provided by https://godoc.org/gopkg.in/go-playground/validator.v9
*/
package radosgwadmin
