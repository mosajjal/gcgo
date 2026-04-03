package kms

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	cloudkms "google.golang.org/api/cloudkms/v1"
	"google.golang.org/api/option"
)

// KeyRing holds key ring metadata.
type KeyRing struct {
	Name       string `json:"name"`
	CreateTime string `json:"create_time"`
}

// CryptoKey holds crypto key metadata.
type CryptoKey struct {
	Name            string `json:"name"`
	Purpose         string `json:"purpose"`
	CreateTime      string `json:"create_time"`
	RotationPeriod  string `json:"rotation_period,omitempty"`
	PrimaryVersion  string `json:"primary_version,omitempty"`
	ProtectionLevel string `json:"protection_level,omitempty"`
}

// CryptoKeyVersion holds key version metadata.
type CryptoKeyVersion struct {
	Name            string `json:"name"`
	State           string `json:"state"`
	ProtectionLevel string `json:"protection_level"`
	Algorithm       string `json:"algorithm"`
	CreateTime      string `json:"create_time"`
}

// AsymmetricSignResult holds asymmetric sign response metadata.
type AsymmetricSignResult struct {
	Name            string `json:"name"`
	ProtectionLevel string `json:"protection_level"`
	Signature       []byte `json:"signature"`
}

// Client defines KMS operations.
type Client interface {
	ListKeyRings(ctx context.Context, project, location string) ([]*KeyRing, error)
	CreateKeyRing(ctx context.Context, project, location, keyRingID string) (*KeyRing, error)
	DescribeKeyRing(ctx context.Context, name string) (*KeyRing, error)
	GetKeyRingPolicy(ctx context.Context, name string) (*cloudkms.Policy, error)
	SetKeyRingPolicy(ctx context.Context, name string, policy *cloudkms.Policy) (*cloudkms.Policy, error)
	TestKeyRingPermissions(ctx context.Context, name string, permissions []string) ([]string, error)
	ListKeys(ctx context.Context, keyRingName string) ([]*CryptoKey, error)
	CreateKey(ctx context.Context, keyRingName, keyID, purpose string) (*CryptoKey, error)
	DescribeKey(ctx context.Context, name string) (*CryptoKey, error)
	GetKeyPolicy(ctx context.Context, name string) (*cloudkms.Policy, error)
	SetKeyPolicy(ctx context.Context, name string, policy *cloudkms.Policy) (*cloudkms.Policy, error)
	TestKeyPermissions(ctx context.Context, name string, permissions []string) ([]string, error)
	ListKeyVersions(ctx context.Context, keyName string) ([]*CryptoKeyVersion, error)
	CreateKeyVersion(ctx context.Context, keyName string) (*CryptoKeyVersion, error)
	UpdatePrimaryVersion(ctx context.Context, keyName, versionName string) (*CryptoKey, error)
	DescribeKeyVersion(ctx context.Context, name string) (*CryptoKeyVersion, error)
	DestroyKeyVersion(ctx context.Context, name string) error
	EnableKeyVersion(ctx context.Context, name string) error
	DisableKeyVersion(ctx context.Context, name string) error
	AsymmetricSign(ctx context.Context, versionName string, data []byte) (*AsymmetricSignResult, error)
	Encrypt(ctx context.Context, keyName string, plaintext []byte) ([]byte, error)
	Decrypt(ctx context.Context, keyName string, ciphertext []byte) ([]byte, error)
}

type gcpClient struct {
	kms *cloudkms.Service
}

// NewClient creates a Client backed by the real GCP KMS API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := cloudkms.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create kms client: %w", err)
	}
	return &gcpClient{kms: svc}, nil
}

func (c *gcpClient) ListKeyRings(ctx context.Context, project, location string) ([]*KeyRing, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	call := c.kms.Projects.Locations.KeyRings.List(parent)

	var rings []*KeyRing
	if err := call.Pages(ctx, func(resp *cloudkms.ListKeyRingsResponse) error {
		for _, kr := range resp.KeyRings {
			rings = append(rings, keyRingFromAPI(kr))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list key rings: %w", err)
	}
	return rings, nil
}

func (c *gcpClient) CreateKeyRing(ctx context.Context, project, location, keyRingID string) (*KeyRing, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	kr, err := c.kms.Projects.Locations.KeyRings.Create(parent, &cloudkms.KeyRing{}).KeyRingId(keyRingID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create key ring %s: %w", keyRingID, err)
	}
	return keyRingFromAPI(kr), nil
}

func (c *gcpClient) DescribeKeyRing(ctx context.Context, name string) (*KeyRing, error) {
	kr, err := c.kms.Projects.Locations.KeyRings.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get key ring %s: %w", name, err)
	}
	return keyRingFromAPI(kr), nil
}

func (c *gcpClient) GetKeyRingPolicy(ctx context.Context, name string) (*cloudkms.Policy, error) {
	policy, err := c.kms.Projects.Locations.KeyRings.GetIamPolicy(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get key ring policy %s: %w", name, err)
	}
	return policy, nil
}

func (c *gcpClient) SetKeyRingPolicy(ctx context.Context, name string, policy *cloudkms.Policy) (*cloudkms.Policy, error) {
	updated, err := c.kms.Projects.Locations.KeyRings.SetIamPolicy(name, &cloudkms.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("set key ring policy %s: %w", name, err)
	}
	return updated, nil
}

func (c *gcpClient) TestKeyRingPermissions(ctx context.Context, name string, permissions []string) ([]string, error) {
	resp, err := c.kms.Projects.Locations.KeyRings.TestIamPermissions(name, &cloudkms.TestIamPermissionsRequest{Permissions: permissions}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("test key ring permissions %s: %w", name, err)
	}
	return resp.Permissions, nil
}

func (c *gcpClient) ListKeys(ctx context.Context, keyRingName string) ([]*CryptoKey, error) {
	call := c.kms.Projects.Locations.KeyRings.CryptoKeys.List(keyRingName)

	var keys []*CryptoKey
	if err := call.Pages(ctx, func(resp *cloudkms.ListCryptoKeysResponse) error {
		for _, k := range resp.CryptoKeys {
			keys = append(keys, cryptoKeyFromAPI(k))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list crypto keys: %w", err)
	}
	return keys, nil
}

func (c *gcpClient) CreateKey(ctx context.Context, keyRingName, keyID, purpose string) (*CryptoKey, error) {
	p := "ENCRYPT_DECRYPT"
	switch purpose {
	case "asymmetric-signing":
		p = "ASYMMETRIC_SIGN"
	case "asymmetric-encryption":
		p = "ASYMMETRIC_DECRYPT"
	case "raw-encrypt-decrypt":
		p = "RAW_ENCRYPT_DECRYPT"
	}

	k, err := c.kms.Projects.Locations.KeyRings.CryptoKeys.Create(keyRingName, &cloudkms.CryptoKey{
		Purpose: p,
	}).CryptoKeyId(keyID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create crypto key %s: %w", keyID, err)
	}
	return cryptoKeyFromAPI(k), nil
}

func (c *gcpClient) DescribeKey(ctx context.Context, name string) (*CryptoKey, error) {
	k, err := c.kms.Projects.Locations.KeyRings.CryptoKeys.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get crypto key %s: %w", name, err)
	}
	return cryptoKeyFromAPI(k), nil
}

func (c *gcpClient) GetKeyPolicy(ctx context.Context, name string) (*cloudkms.Policy, error) {
	policy, err := c.kms.Projects.Locations.KeyRings.CryptoKeys.GetIamPolicy(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get crypto key policy %s: %w", name, err)
	}
	return policy, nil
}

func (c *gcpClient) SetKeyPolicy(ctx context.Context, name string, policy *cloudkms.Policy) (*cloudkms.Policy, error) {
	updated, err := c.kms.Projects.Locations.KeyRings.CryptoKeys.SetIamPolicy(name, &cloudkms.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("set crypto key policy %s: %w", name, err)
	}
	return updated, nil
}

func (c *gcpClient) TestKeyPermissions(ctx context.Context, name string, permissions []string) ([]string, error) {
	resp, err := c.kms.Projects.Locations.KeyRings.CryptoKeys.TestIamPermissions(name, &cloudkms.TestIamPermissionsRequest{Permissions: permissions}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("test crypto key permissions %s: %w", name, err)
	}
	return resp.Permissions, nil
}

func (c *gcpClient) ListKeyVersions(ctx context.Context, keyName string) ([]*CryptoKeyVersion, error) {
	call := c.kms.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.List(keyName)

	var versions []*CryptoKeyVersion
	if err := call.Pages(ctx, func(resp *cloudkms.ListCryptoKeyVersionsResponse) error {
		for _, v := range resp.CryptoKeyVersions {
			versions = append(versions, keyVersionFromAPI(v))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list key versions: %w", err)
	}
	return versions, nil
}

func (c *gcpClient) CreateKeyVersion(ctx context.Context, keyName string) (*CryptoKeyVersion, error) {
	v, err := c.kms.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.Create(keyName, &cloudkms.CryptoKeyVersion{}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create key version for %s: %w", keyName, err)
	}
	return keyVersionFromAPI(v), nil
}

func (c *gcpClient) UpdatePrimaryVersion(ctx context.Context, keyName, versionName string) (*CryptoKey, error) {
	key, err := c.kms.Projects.Locations.KeyRings.CryptoKeys.UpdatePrimaryVersion(keyName, &cloudkms.UpdateCryptoKeyPrimaryVersionRequest{
		CryptoKeyVersionId: keyVersionID(versionName),
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("set primary version for %s: %w", keyName, err)
	}
	return cryptoKeyFromAPI(key), nil
}

func (c *gcpClient) DescribeKeyVersion(ctx context.Context, name string) (*CryptoKeyVersion, error) {
	v, err := c.kms.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get key version %s: %w", name, err)
	}
	return keyVersionFromAPI(v), nil
}

func (c *gcpClient) DestroyKeyVersion(ctx context.Context, name string) error {
	if _, err := c.kms.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.Destroy(name, &cloudkms.DestroyCryptoKeyVersionRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("destroy key version %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) EnableKeyVersion(ctx context.Context, name string) error {
	if _, err := c.kms.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.Patch(name, &cloudkms.CryptoKeyVersion{State: "ENABLED"}).UpdateMask("state").Context(ctx).Do(); err != nil {
		return fmt.Errorf("enable key version %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) DisableKeyVersion(ctx context.Context, name string) error {
	if _, err := c.kms.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.Patch(name, &cloudkms.CryptoKeyVersion{State: "DISABLED"}).UpdateMask("state").Context(ctx).Do(); err != nil {
		return fmt.Errorf("disable key version %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) AsymmetricSign(ctx context.Context, versionName string, data []byte) (*AsymmetricSignResult, error) {
	resp, err := c.kms.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.AsymmetricSign(versionName, &cloudkms.AsymmetricSignRequest{
		Data: base64.StdEncoding.EncodeToString(data),
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("asymmetric sign %s: %w", versionName, err)
	}
	signature, err := base64.StdEncoding.DecodeString(resp.Signature)
	if err != nil {
		return nil, fmt.Errorf("decode asymmetric signature: %w", err)
	}
	return &AsymmetricSignResult{
		Name:            resp.Name,
		ProtectionLevel: resp.ProtectionLevel,
		Signature:       signature,
	}, nil
}

func (c *gcpClient) Encrypt(ctx context.Context, keyName string, plaintext []byte) ([]byte, error) {
	resp, err := c.kms.Projects.Locations.KeyRings.CryptoKeys.Encrypt(keyName, &cloudkms.EncryptRequest{
		Plaintext: base64.StdEncoding.EncodeToString(plaintext),
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(resp.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}
	return ciphertext, nil
}

func (c *gcpClient) Decrypt(ctx context.Context, keyName string, ciphertext []byte) ([]byte, error) {
	resp, err := c.kms.Projects.Locations.KeyRings.CryptoKeys.Decrypt(keyName, &cloudkms.DecryptRequest{
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	plaintext, err := base64.StdEncoding.DecodeString(resp.Plaintext)
	if err != nil {
		return nil, fmt.Errorf("decode plaintext: %w", err)
	}
	return plaintext, nil
}

func keyRingFromAPI(kr *cloudkms.KeyRing) *KeyRing {
	return &KeyRing{
		Name:       kr.Name,
		CreateTime: kr.CreateTime,
	}
}

func cryptoKeyFromAPI(k *cloudkms.CryptoKey) *CryptoKey {
	ck := &CryptoKey{
		Name:           k.Name,
		Purpose:        k.Purpose,
		CreateTime:     k.CreateTime,
		RotationPeriod: k.RotationPeriod,
	}
	if k.VersionTemplate != nil {
		ck.ProtectionLevel = k.VersionTemplate.ProtectionLevel
	}
	if k.Primary != nil {
		ck.PrimaryVersion = k.Primary.Name
		if ck.ProtectionLevel == "" {
			ck.ProtectionLevel = k.Primary.ProtectionLevel
		}
	}
	return ck
}

func keyVersionFromAPI(v *cloudkms.CryptoKeyVersion) *CryptoKeyVersion {
	return &CryptoKeyVersion{
		Name:            v.Name,
		State:           v.State,
		ProtectionLevel: v.ProtectionLevel,
		Algorithm:       v.Algorithm,
		CreateTime:      v.CreateTime,
	}
}

func keyVersionID(name string) string {
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		return name[idx+1:]
	}
	return name
}
