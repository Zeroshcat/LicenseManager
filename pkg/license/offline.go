// Package license 提供许可证生成和验证功能
package license

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"os"
	"time"
	
	"licensemanager/internal/crypto"
)

// OfflineVerifier 离线验证器
// 完全本地验证，不需要网络连接
type OfflineVerifier struct {
	publicKey *rsa.PublicKey // RSA公钥（用于验证签名）
	aesKey    []byte        // AES密钥（用于解密）
}

// NewOfflineVerifier 创建离线验证器
// 参数：
//   - publicKeyPEM: RSA公钥（PEM格式），用于验证许可证签名
//   - aesKey: AES密钥（32字节），用于解密许可证
// 返回值：
//   - *OfflineVerifier: 离线验证器实例
func NewOfflineVerifier(publicKeyPEM []byte, aesKey []byte) *OfflineVerifier {
	// 解码公钥
	publicKey, _ := crypto.DecodePublicKey(publicKeyPEM)
	
	return &OfflineVerifier{
		publicKey: publicKey,
		aesKey:    aesKey,
	}
}

// Verify 验证离线许可证
// 参数：
//   - licenseKey: 许可证密钥（base64编码）
//   - deviceID: 设备ID
// 返回值：
//   - *VerifyResult: 验证结果
//   - error: 验证过程中的错误
func (v *OfflineVerifier) Verify(licenseKey string, deviceID string) (*VerifyResult, error) {
	// 解码许可证
	license, err := v.decodeLicense(licenseKey)
	if err != nil {
		return nil, ErrInvalidLicense
	}
	
	// 检查设备ID
	if license.DeviceID != deviceID {
		return nil, ErrDeviceMismatch
	}
	
	// 检查是否过期
	now := time.Now()
	expired := now.After(license.ExpiryDate)
	
	result := &VerifyResult{
		Valid:       !expired,
		Expired:     expired,
		ExpiryDate:  license.ExpiryDate,
		DeviceID:    license.DeviceID,
		LicenseType: string(license.LicenseType),
		Message:     "Offline verification",
	}
	
	if expired {
		result.Message = "License expired"
		return result, ErrExpiredLicense
	}
	
	return result, nil
}

// decodeLicense 解码许可证密钥
// 参数：
//   - licenseKey: base64编码的许可证密钥
// 返回值：
//   - *License: 许可证对象
//   - error: 解码过程中的错误
func (v *OfflineVerifier) decodeLicense(licenseKey string) (*License, error) {
	// Base64解码
	licenseData, err := base64.StdEncoding.DecodeString(licenseKey)
	if err != nil {
		return nil, err
	}
	
	// 提取签名和密文（RSA-4096签名长度为512字节）
	signatureSize := 512
	if len(licenseData) < signatureSize {
		return nil, ErrInvalidLicense
	}
	
	signature := licenseData[:signatureSize]
	encryptedData := licenseData[signatureSize:]
	
	// 验证签名
	valid, err := crypto.VerifySignature(encryptedData, signature, v.publicKey)
	if err != nil || !valid {
		return nil, ErrInvalidLicense
	}
	
	// 使用AES解密
	jsonData, err := crypto.DecryptAES(encryptedData, v.aesKey)
	if err != nil {
		return nil, ErrInvalidLicense
	}
	
	// 反序列化
	var license License
	if err := json.Unmarshal(jsonData, &license); err != nil {
		return nil, ErrInvalidLicense
	}
	
	return &license, nil
}

// LoadLicenseFromFile 从文件加载许可证
// 参数：
//   - filepath: 许可证文件路径
// 返回值：
//   - string: 许可证密钥
//   - error: 加载过程中的错误
func LoadLicenseFromFile(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", ErrLicenseNotFound
	}
	return string(data), nil
}

