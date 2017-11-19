package radosgwadmin

import (
	"context"
)

// UserCreateRequest - describes what to do in a user create operation.
type UserCreateRequest struct {
	UID         string    `url:"uid" validate:"required"`
	DisplayName string    `url:"display-name" validate:"required"`
	Email       string    `url:"email,omitempty" validate:"omitempty,email"`
	KeyType     string    `url:"key-type,omitempty" validate:"omitempty,eq=swift|eq=s3"`
	AccessKey   string    `url:"access-key,omitempty"`
	SecretKey   string    `url:"secret-key,omitempty"`
	UserCaps    []UserCap `url:"user-caps,omitempty,semicolon" validate:"omitempty,dive"`
	Tenant      string    `url:"tenant,omitempty"`
	GenerateKey *bool     `url:"generate-key,omitempty"` // This defaults to true, preserving that behavior
	MaxBuckets  int       `url:"max-buckets,omitempty"`
	Suspended   bool      `url:"suspended,omitempty"`
}

// UserModifyRequest - modify user request type.
type UserModifyRequest struct {
	UID         string    `url:"uid" validate:"required"`
	DisplayName string    `url:"display-name,omitempty"`
	Email       string    `url:"email,omitempty"`
	KeyType     string    `url:"key-type,omitempty" validate:"omitempty,eq=swift|eq=s3"`
	AccessKey   string    `url:"access-key,omitempty"`
	SecretKey   string    `url:"secret-key,omitempty"`
	UserCaps    []UserCap `url:"user-caps,omitempty,semicolon" validate:"omitempty,dive"`
	GenerateKey bool      `url:"generate-key,omitempty"` // This defaults to false, preserving that behavior
	MaxBuckets  int       `url:"max-buckets,omitempty"`
	Suspended   bool      `url:"suspended,omitempty"`
}

type userInfoRequest struct {
	UID   string `url:"uid" validate:"required"`
	Stats bool   `url:"stats,omitempty"`
}

type userDeleteRequest struct {
	UID       string `url:"uid" validate:"required"`
	PurgeData bool   `url:"purge-data"`
}

// UserCapsRequest - this is passed to CapsAdd() and CapsRm()
type UserCapsRequest struct {
	UID      string    `url:"uid" validate:"required"`
	UserCaps []UserCap `url:"user-caps,semicolon" validate:"required,dive"`
}

// UserInfoResponse - response from a user info request.
type UserInfoResponse struct {
	Tenant      string     `json:"tenant"`
	UserID      string     `json:"user_id"`
	DisplayName string     `json:"display_name"`
	Email       string     `json:"email"`
	Suspended   int        `json:"suspended"` // should be bool
	MaxBuckets  int        `json:"max_buckets"`
	SubUsers    []SubUser  `json:"subusers"`
	Keys        []UserKey  `json:"keys"`
	SwiftKeys   []SwiftKey `json:"swift_keys"`
	Caps        []UserCap  `json:"caps"`
	// Stats is returned if the stats flag is passed to the user info request.
	Stats *UserStats `json:"stats"`
}

//UserStats - statistics for a user
type UserStats struct {
	Size           int `json:"size"`
	SizeActual     int `json:"size_actual"`
	SizeUtilized   int `json:"size_utilized"`
	SizeKB         int `json:"size_kb"`
	SizeKBActual   int `json:"size_kb_actual"`
	SizeKBUtilized int `json:"size_kb_utilized"`
	NumObjects     int `json:"num_objects"`
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

// UserCap - desribes user capabilities / permissions.
type UserCap struct {
	Type       string `json:"type" validate:"required,eq=users|eq=buckets|eq=metadata|eq=usage|eq=zone"`
	Permission string `json:"perm" validate:"required,eq=*|eq=read|eq=write|eq=read0x2Cwrite"`
}

// String - Implement Stringer
func (uc UserCap) String() string {
	return uc.Type + "=" + uc.Permission
}

// SubUserCreateModifyRequest - Create or modify sub user request.
type SubUserCreateModifyRequest struct {
	UID            string `url:"uid" validate:"required"`
	SubUser        string `url:"subuser" validate:"required"`
	SecretKey      string `url:"secret-key,omitempty"`
	KeyType        string `url:"key-type,omitempty" validate:"omitempty,eq=s3|eq=swift"`
	Access         string `url:"access,omitempty" validate:"omitempty,eq=read|eq=write|eq=readwrite|eq=full"`
	GenerateSecret bool   `url:"generate-secret,omitempty"`
}

// SubUserRmRequest - if PurgeKeys is nil, defaults to true
type SubUserRmRequest struct {
	UID       string `url:"uid" validate:"required"`
	SubUser   string `url:"subuser" validate:"required"`
	PurgeKeys *bool  `url:"purge-keys,omitempty"` // Default true
}

// UserInfo - get user information about uid.  If stats is true, then return
// quota statistics.  This will return a not found error if no statistics
// are available, even if the user exists.
func (aa *AdminAPI) UserInfo(ctx context.Context, uid string, stats bool) (*UserInfoResponse, error) {
	uir := &userInfoRequest{uid, stats}
	resp := &UserInfoResponse{}

	err := aa.Get(ctx, "/user", uir, resp)
	return resp, err
}

// UserCreate - create a user described by cur.
func (aa *AdminAPI) UserCreate(ctx context.Context, cur *UserCreateRequest) (*UserInfoResponse, error) {
	resp := &UserInfoResponse{}
	err := aa.Put(ctx, "/user", cur, nil, resp)
	return resp, err
}

// UserRm - delete user uid
func (aa *AdminAPI) UserRm(ctx context.Context, uid string, purge bool) error {
	udr := &userDeleteRequest{uid, purge}
	return aa.Delete(ctx, "/user", udr, nil)
}

// UserModify - modify a user described by umr
func (aa *AdminAPI) UserModify(ctx context.Context, umr *UserModifyRequest) (*UserInfoResponse, error) {
	resp := &UserInfoResponse{}
	err := aa.Post(ctx, "/user", umr, nil, resp)
	return resp, err
}

// SubUserCreate - create a subuser
func (aa *AdminAPI) SubUserCreate(ctx context.Context, sucr *SubUserCreateModifyRequest) ([]SubUser, error) {
	resp := []SubUser{}
	err := aa.Put(ctx, "/user?subuser", sucr, nil, &resp)
	return resp, err
}

// SubUserModify - modify a subuser
func (aa *AdminAPI) SubUserModify(ctx context.Context, sucr *SubUserCreateModifyRequest) ([]SubUser, error) {
	resp := []SubUser{}
	err := aa.Post(ctx, "/user?subuser", sucr, nil, &resp)
	return resp, err
}

// SubUserRm - delete a subuser
func (aa *AdminAPI) SubUserRm(ctx context.Context, surm *SubUserRmRequest) error {
	return aa.Delete(ctx, "/user?subuser", surm, nil)
}

// CapsAdd - Add capabilities / permissions.  Returns the new effective capabilities.
//
// Note - capabilities are additive.  This will only ever make a user's permissions
// more permissive.  As an example, if the user has metadata permission of *, calling
// this with metadata set to read will have no effect.  On the other hand, if a user's
// permission was read, and CapsAdd was called with write, the new effective permission
// would be read + write (*).  To remove permissions, you must call CapsRm(), which is
// subtractive.
func (aa *AdminAPI) CapsAdd(ctx context.Context, ucr *UserCapsRequest) ([]UserCap, error) {
	resp := []UserCap{}
	err := aa.Put(ctx, "/user?caps", ucr, nil, &resp)
	return resp, err
}

// CapsRm - Remove capabilities / permissions.  Returns the new effective capabilities.
// See notes for CapsAdd().
func (aa *AdminAPI) CapsRm(ctx context.Context, ucr *UserCapsRequest) ([]UserCap, error) {
	resp := []UserCap{}
	err := aa.Delete(ctx, "/user?caps", ucr, &resp)
	return resp, err
}
