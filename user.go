package radosgwadmin

import (
	"context"
)

type userInfoRequest struct {
	UID string `url:"uid"`
}

type userDeleteRequest struct {
	UID       string `url:"uid"`
	PurgeData bool   `url:"purge-data"`
}

// UserInfoResponse - respone from a user info request.
type UserInfoResponse struct {
	UserID      string     `json:"user_id"`
	DisplayName string     `json:"display_name"`
	Email       string     `json:"email"`
	Suspended   int        `json:"suspended"` // should be bool
	MaxBuckets  int        `json:"max_buckets"`
	SubUsers    []SubUser  `json:"subusers"`
	Keys        []UserKey  `json:"keys"`
	SwiftKeys   []SwiftKey `json:"swift_keys"`
	Caps        UserCaps   `json:"caps"`
}

// UserCreateRequest - describes what to do in a user create operation.
type UserCreateRequest struct {
	UID         string   `url:"uid" validate:"required"`
	DisplayName string   `url:"display-name"`
	Email       string   `url:"email,omitempty" validate:"omitempty,email"`
	KeyType     string   `url:"key-type,omitempty" validate:"omitempty,eq=swift|eq=s3"`
	AccessKey   string   `url:"access-key,omitempty"`
	SecretKey   string   `url:"secret-key,omitempty"`
	UserCaps    UserCaps `url:"user-caps,omitempty,semicolon"`
	GenerateKey *bool    `url:"generate-key,omitempty"` // This defaults to true, preserving that behavior
	MaxBuckets  int      `url:"max-buckets,omitempty"`
	Suspended   bool     `url:"suspended,omitempty"`
}

// UserModifyRequest - modify user request type.
type UserModifyRequest struct {
	UID         string   `url:"uid" validate:"required"`
	DisplayName string   `url:"display-name,omitempty"`
	Email       string   `url:"email,omitempty"`
	KeyType     string   `url:"key-type,omitempty" validate:"omitempty,eq=swift|eq=s3"`
	AccessKey   string   `url:"access-key,omitempty"`
	SecretKey   string   `url:"secret-key,omitempty"`
	UserCaps    UserCaps `url:"user-caps,omitempty,semicolon"`
	GenerateKey bool     `url:"generate-key,omitempty"` // This defaults to false, preserving that behavior
	MaxBuckets  int      `url:"max-buckets,omitempty"`
	Suspended   bool     `url:"suspended,omitempty"`
}

// SubUser - describes a subuser
type SubUser struct {
	ID          string `json:"id"`
	Permissions string `json:"permissions"`
}

// UserKey - user key information.
type UserKey struct {
	User      string `json:"user"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

// SwiftKey - swift key information
type SwiftKey struct {
	User      string `json:"user"`
	SecretKey string `json:"secret_key"`
}

// UserCapability - desribes user capabilities / permissions.
type UserCapability struct {
	Type       string `json:"type" validate:"users|buckets|metadata|usage|zone"`
	Permission string `json:"perm" enum:"*|read|write|read,write"`
}

// Implement Stringer
func (uc UserCapability) String() string {
	return uc.Type + "=" + uc.Permission
}

// SubUserCreateRequest - Create sub user request.
type SubUserCreateRequest struct {
	UID            string `url:"uid" validate:"required"`
	SubUser        string `url:"subuser" validate:"required"`
	SecretKey      string `url:"secret-key,omitempty"`
	KeyType        string `url:"key-type,omitempty"`
	Access         string `url:"access,omitempty" validate:"omitempty,eq=read|eq=write|eq=readwrite|eq=full"`
	GenerateSecret bool   `url:"generate-secret,omitempty"`
}

// UserCaps - list of UserCapability
type UserCaps []UserCapability

// UserInfo - get user information about uid.
func (aa *AdminAPI) UserInfo(ctx context.Context, uid string) (*UserInfoResponse, error) {
	uir := &userInfoRequest{uid}
	resp := &UserInfoResponse{}

	err := aa.get(ctx, "/user", uir, resp)
	return resp, err
}

// UserCreate - create a user described by cur.
func (aa *AdminAPI) UserCreate(ctx context.Context, cur *UserCreateRequest) (*UserInfoResponse, error) {
	resp := &UserInfoResponse{}
	err := aa.put(ctx, "/user", cur, nil, resp)
	return resp, err
}

// UserRm - delete user uid
func (aa *AdminAPI) UserRm(ctx context.Context, uid string, purge bool) error {
	udr := &userDeleteRequest{uid, purge}
	return aa.delete(ctx, "/user", udr, nil)
}

// UserUpdate - update a user described by umr
func (aa *AdminAPI) UserUpdate(ctx context.Context, umr *UserModifyRequest) (*UserInfoResponse, error) {
	resp := &UserInfoResponse{}
	err := aa.post(ctx, "/user", umr, nil, resp)
	return resp, err
}

// SubUserCreate - create a subuser
func (aa *AdminAPI) SubUserCreate(ctx context.Context, sucr *SubUserCreateRequest) ([]SubUser, error) {
	resp := []SubUser{}
	err := aa.put(ctx, "/user?subuser", sucr, nil, &resp)
	return resp, err
}
