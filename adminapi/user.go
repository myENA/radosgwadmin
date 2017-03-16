package adminapi

type UserInfoRequest struct {
	Uid string `url:"uid"`
}

type UserInfoResponse struct {
	UserId      string             `json:"user_id"`
	DisplayName string             `json:"display_name"`
	Email       string             `json:"email"`
	Suspended   int                `json:"suspended"` // should be bool
	MaxBuckets  int                `json:"max_buckets"`
	SubUsers    []SubUser          `json:"subusers"`
	Keys        []UserKey          `json:"keys"`
	SwiftKeys   []SwiftKey         `json:"swift_keys"`
	Caps        []UserCapabilities `json:"caps"`
}

type CreateUserRequest struct {
	Uid         string `json:"uid"`
	DisplayName string `json:"display-name"`
	Email       string `json:"email,omitempty"`
	KeyType     string `json:"key-type,omitempty" enum:"swift|s3|"`
	AccessKey   string `json:"access-key,omitempty"`
	SecretKey   string `json:"secret-key,omitempty"`
	UserCaps    string `json:"user-caps"`
	GenerateKey *bool  `json:"generate-key,omitempty"` // This defaults to true, preserving that behavior
	MaxBuckets  int    `json:"max-buckets,omitempty"`
	Suspended   bool   `json:"suspended,omitempty"`
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

type UserCapabilities struct {
	Type       string `json:"type"`
	Permission string `json:"perm"`
}

func (aa *AdminApi) GetUserInfo(uid string) (*UserInfoResponse, error) {
	uir := &UserInfoRequest{uid}
	resp := &UserInfoResponse{}

	err := aa.Get("/user", uir, resp)
	return resp, err
}
