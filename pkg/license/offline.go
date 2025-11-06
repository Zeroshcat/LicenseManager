// Package license 提供许可证生成和验证功能
package license

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Zeroshcat/LicenseManager/internal/crypto"
)

// OfflineVerifier 离线验证器
// 完全本地验证，不需要网络连接
type OfflineVerifier struct {
	publicKey *rsa.PublicKey // RSA公钥（用于验证签名）
	aesKey    []byte         // AES密钥（用于解密）
}

// NewOfflineVerifier 创建离线验证器
// 参数：
//   - publicKeyPEM: RSA公钥（PEM格式），用于验证许可证签名
//   - aesKey: AES密钥（32字节），用于解密许可证
//
// 返回值：
//   - *OfflineVerifier: 离线验证器实例
//   - error: 创建过程中的错误
func NewOfflineVerifier(publicKeyPEM []byte, aesKey []byte) (*OfflineVerifier, error) {
	// 解码公钥
	publicKey, err := crypto.DecodePublicKey(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	// 验证AES密钥长度
	if len(aesKey) != 32 {
		return nil, fmt.Errorf("AES key must be 32 bytes, got %d bytes", len(aesKey))
	}

	return &OfflineVerifier{
		publicKey: publicKey,
		aesKey:    aesKey,
	}, nil
}

// Verify 验证离线许可证
// 参数：
//   - licenseKey: 许可证密钥（base64编码）
//   - deviceID: 设备ID
//
// 返回值：
//   - *VerifyResult: 验证结果
//   - error: 验证过程中的错误
func (v *OfflineVerifier) Verify(licenseKey string, deviceID string) (*VerifyResult, error) {
	// 解码许可证
	license, err := v.decodeLicense(licenseKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode license: %w", err)
	}

	// 防御性检查：license 不应该为 nil
	if license == nil {
		return nil, fmt.Errorf("decoded license is nil")
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
//
// 返回值：
//   - *License: 许可证对象
//   - error: 解码过程中的错误
func (v *OfflineVerifier) decodeLicense(licenseKey string) (*License, error) {
	// Base64解码
	licenseData, err := base64.StdEncoding.DecodeString(licenseKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// 提取签名和密文（RSA-4096签名长度为512字节）
	signatureSize := 512
	if len(licenseData) < signatureSize {
		return nil, ErrInvalidLicense
	}

	signature := licenseData[:signatureSize]
	encryptedData := licenseData[signatureSize:]

	// 检查公钥是否为 nil（防御性编程）
	if v.publicKey == nil {
		return nil, fmt.Errorf("public key is nil")
	}

	// 验证签名
	valid, err := crypto.VerifySignature(encryptedData, signature, v.publicKey)
	if err != nil || !valid {
		return nil, ErrInvalidLicense
	}

	// 使用AES解密
	jsonData, err := crypto.DecryptAES(encryptedData, v.aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AES: %w", err)
	}

	// 反序列化
	var license License
	if err := json.Unmarshal(jsonData, &license); err != nil {
		return nil, fmt.Errorf("failed to unmarshal license: %w", err)
	}

	return &license, nil
}

// LoadLicenseFromFile 从文件加载许可证
// 参数：
//   - filepath: 许可证文件路径
//
// 返回值：
//   - string: 许可证密钥
//   - error: 加载过程中的错误
func LoadLicenseFromFile(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", ErrLicenseNotFound
	}

	// 清理许可证密钥：去除换行符、空格等
	licenseKey := strings.TrimSpace(string(data))
	licenseKey = strings.ReplaceAll(licenseKey, "\n", "")
	licenseKey = strings.ReplaceAll(licenseKey, "\r", "")
	licenseKey = strings.ReplaceAll(licenseKey, " ", "")

	return licenseKey, nil
}
