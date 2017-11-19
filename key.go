package radosgwadmin

import (
	"context"
)

// KeyCreateRequest - Create or modify a key.
type KeyCreateRequest struct {
	UID         string `url:"uid" validate:"required"`
	SubUser     string `url:"subuser,omitempty"`
	AccessKey   string `url:"access-key,omitempty"`
	SecretKey   string `url:"secret-key,omitempty"`
	KeyType     string `url:"key-type,omitempty" validate:"omitempty,eq=s3|eq=swift"`
	GenerateKey *bool  `url:"generate-key,omitempty"` // defaults to true
}

// KeyRmRequest - Create or modify a key.
type KeyRmRequest struct {
	AccessKey string `url:"access-key" validate:"required"`
	UID       string `url:"uid,omitempty"`
	SubUser   string `url:"subuser,omitempty"`
	KeyType   string `url:"key-type,omitempty" validate:"omitempty,eq=s3|eq=swift"`
}

// KeyCreate - Create a key
//
// Create a new key. If a subuser is specified then by default created keys will
// be swift type. If only one of access-key or secret-key is provided the committed
// key will be automatically generated, that is if only SecretKey is specified then
// AccessKey will be automatically generated. By default, a generated key is added
// to the keyring without replacing an existing key pair. If access-key is specified
// and refers to an existing key owned by the user then it will be modified.
// The response is a container listing all keys of the same type as the key created.
// Note that when creating a swift key, specifying the option AccessKey will have no
// effect. Additionally, only one swift key may be held by each user or subuser.
func (aa *AdminAPI) KeyCreate(ctx context.Context, kcr *KeyCreateRequest) ([]UserKey, error) {
	resp := []UserKey{}
	err := aa.Put(ctx, "/user?key", kcr, nil, &resp)
	return resp, err
}

// KeyRm - delete an existing key.
//
// Key type is optional, but required to remove a swift key.
func (aa *AdminAPI) KeyRm(ctx context.Context, krr *KeyRmRequest) error {
	return aa.Delete(ctx, "/user?key", krr, nil)
}
