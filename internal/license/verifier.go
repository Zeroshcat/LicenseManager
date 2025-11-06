// Package license 提供许可证验证功能
package license

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"time"
	
	"licensemanager/internal/crypto"
	"licensemanager/pkg/license"
)

// Verifier 许可证验证器
type Verifier struct {
	publicKey *rsa.PublicKey // RSA公钥（用于验证签名）
	aesKey    []byte         // AES密钥（用于解密）
}

// NewVerifier 创建许可证验证器
// 参数：
//   - publicKey: RSA公钥
//   - aesKey: AES密钥（32字节）
// 返回值：
//   - *Verifier: 许可证验证器实例
func NewVerifier(publicKey *rsa.PublicKey, aesKey []byte) *Verifier {
	return &Verifier{
		publicKey: publicKey,
		aesKey:    aesKey,
	}
}

// Verify 验证许可证
// 参数：
//   - licenseKey: base64编码的许可证密钥
//   - deviceID: 设备ID
// 返回值：
//   - *license.VerifyResult: 验证结果
//   - error: 验证过程中的错误
func (v *Verifier) Verify(licenseKey string, deviceID string) (*license.VerifyResult, error) {
	// Base64解码
	licenseData, err := base64.StdEncoding.DecodeString(licenseKey)
	if err != nil {
		return nil, license.ErrInvalidLicense
	}
	
	// 提取签名和密文（RSA-4096签名长度为512字节）
	signatureSize := 512
	if len(licenseData) < signatureSize {
		return nil, license.ErrInvalidLicense
	}
	
	signature := licenseData[:signatureSize]
	encryptedData := licenseData[signatureSize:]
	
	// 验证签名
	valid, err := crypto.VerifySignature(encryptedData, signature, v.publicKey)
	if err != nil || !valid {
		return nil, license.ErrInvalidLicense
	}
	
	// 使用AES解密
	jsonData, err := crypto.DecryptAES(encryptedData, v.aesKey)
	if err != nil {
		return nil, license.ErrInvalidLicense
	}
	
	// 反序列化
	var lic license.License
	if err := json.Unmarshal(jsonData, &lic); err != nil {
		return nil, license.ErrInvalidLicense
	}
	
	// 检查设备ID
	if lic.DeviceID != deviceID {
		return nil, license.ErrDeviceMismatch
	}
	
	// 检查是否过期
	now := time.Now()
	expired := now.After(lic.ExpiryDate)
	
	result := &license.VerifyResult{
		Valid:       !expired,
		Expired:     expired,
		ExpiryDate:  lic.ExpiryDate,
		DeviceID:    lic.DeviceID,
		LicenseType: string(lic.LicenseType),
		Message:     "License verified",
	}
	
	if expired {
		result.Message = "License expired"
		return result, license.ErrExpiredLicense
	}
	
	return result, nil
}


