package adminapi

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
	Type       string `json:"type"`
	Permission string `json:"perm"`
}

type UserCaps []UserCapability

func (uc UserCapability) String() string {
	return uc.Type + "=" + uc.Permission
}

func (aa *AdminApi) UserInfo(uid string) (*UserInfoResponse, error) {
	uir := &UserInfoRequest{uid}
	resp := &UserInfoResponse{}

	err := aa.Get("/user", uir, resp)
	return resp, err
}

func (aa *AdminApi) UserCreate(cur *UserCreateRequest) (*UserInfoResponse, error) {
	resp := &UserInfoResponse{}
	err := aa.Put("/user", cur, nil, resp)
	return resp, err
}

func (aa *AdminApi) UserRm(uid string) error {
	udr := &UserDeleteRequest{uid}
	return aa.Delete("/user", udr, nil)
}

func (aa *AdminApi) UserUpdate(umr *UserModifyRequest) (*UserInfoResponse, error) {
	resp := &UserInfoResponse{}
	err := aa.Post("/user", umr, nil, resp)
	return resp, err
}
