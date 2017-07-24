package adminapi

import (
	"context"
)

type UserInfoRequest struct {
	Uid string `url:"uid"`
}

type UserDeleteRequest UserInfoRequest

type UserInfoResponse struct {
	UserId      string     `json:"user_id"`
	DisplayName string     `json:"display_name"`
	Email       string     `json:"email"`
	Suspended   int        `json:"suspended"` // should be bool
	MaxBuckets  int        `json:"max_buckets"`
	SubUsers    []SubUser  `json:"subusers"`
	Keys        []UserKey  `json:"keys"`
	SwiftKeys   []SwiftKey `json:"swift_keys"`
	Caps        UserCaps   `json:"caps"`
}

type UserCreateRequest struct {
	Uid         string   `url:"uid"`
	DisplayName string   `url:"display-name"`
	Email       string   `url:"email,omitempty"`
	KeyType     string   `url:"key-type,omitempty" enum:"swift|s3|"`
	AccessKey   string   `url:"access-key,omitempty"`
	SecretKey   string   `url:"secret-key,omitempty"`
	UserCaps    UserCaps `url:"user-caps,omitempty,semicolon"`
	GenerateKey *bool    `url:"generate-key,omitempty"` // This defaults to true, preserving that behavior
	MaxBuckets  int      `url:"max-buckets,omitempty"`
	Suspended   bool     `url:"suspended,omitempty"`
}

type UserModifyRequest struct {
	Uid         string   `url:"uid"`
	DisplayName string   `url:"display-name"`
	Email       string   `url:"email,omitempty"`
	KeyType     string   `url:"key-type,omitempty" enum:"swift|s3|"`
	AccessKey   string   `url:"access-key,omitempty"`
	SecretKey   string   `url:"secret-key,omitempty"`
	UserCaps    UserCaps `url:"user-caps,omitempty,semicolon"`
	GenerateKey bool     `url:"generate-key,omitempty"` // This defaults to false, preserving that behavior
	MaxBuckets  int      `url:"max-buckets,omitempty"`
	Suspended   bool     `url:"suspended,omitempty"`
}

type SubUser struct {
	Id          string `json:"id"`
	Permissions string `json:"permissions"`
}

type UserKey struct {
	User      string `json:"user"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

type SwiftKey struct {
	User      string `json:"user"`
	SecretKey string `json:"secret_key"`
}

type UserCapability struct {
	Type       string `json:"type" enum:"users|buckets|metadata|usage|zone"`
	Permission string `json:"perm" enum:"*|read|write|read,write"`
}

// Implement Stringer
func (uc UserCapability) String() string {
	return uc.Type + "=" + uc.Permission
}

type SubUserCreateRequest struct {
	UID            string `url:"uid"`
	SubUser        string `url:"subuser"`
	SecretKey      string `url:"secret-key,omitempty"`
	KeyType        string `url:"key-type,omitempty"`
	Access         string `url:"access,omitempty"`
	GenerateSecret bool   `url:"generate-secret,omitempty"`
}

type UserCaps []UserCapability

func (aa *AdminApi) UserInfo(ctx context.Context, uid string) (*UserInfoResponse, error) {
	uir := &UserInfoRequest{uid}
	resp := &UserInfoResponse{}

	err := aa.get(ctx, "/user", uir, resp)
	return resp, err
}

func (aa *AdminApi) UserCreate(ctx context.Context, cur *UserCreateRequest) (*UserInfoResponse, error) {
	resp := &UserInfoResponse{}
	err := aa.put(ctx, "/user", cur, nil, resp)
	return resp, err
}

func (aa *AdminApi) UserRm(ctx context.Context, uid string) error {
	udr := &UserDeleteRequest{uid}
	return aa.delete(ctx, "/user", udr, nil)
}

func (aa *AdminApi) UserUpdate(ctx context.Context, umr *UserModifyRequest) (*UserInfoResponse, error) {
	resp := &UserInfoResponse{}
	err := aa.post(ctx, "/user", umr, nil, resp)
	return resp, err
}

func (aa *AdminApi) SubUserCreate(ctx context.Context, sucr *SubUserCreateRequest) ([]SubUser, error) {
	resp := []SubUser{}
	err := aa.put(ctx, "/user?subuser", sucr, nil, &resp)
	return resp, err
}
