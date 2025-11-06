// Package license 提供许可证生成功能
package license

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"time"
	
	"licensemanager/internal/crypto"
	"licensemanager/pkg/license"
)

// Generator 许可证生成器
type Generator struct {
	privateKey *rsa.PrivateKey // RSA私钥（用于签名）
	aesKey     []byte          // AES密钥（用于加密）
}

// NewGenerator 创建许可证生成器
// 参数：
//   - privateKey: RSA私钥
//   - aesKey: AES密钥（32字节）
// 返回值：
//   - *Generator: 许可证生成器实例
func NewGenerator(privateKey *rsa.PrivateKey, aesKey []byte) *Generator {
	return &Generator{
		privateKey: privateKey,
		aesKey:     aesKey,
	}
}

// Generate 生成许可证
// 参数：
//   - deviceID: 设备ID
//   - licenseType: 许可证类型
//   - expiryDate: 到期时间
//   - features: 功能列表
// 返回值：
//   - string: base64编码的许可证密钥
//   - error: 生成过程中的错误
func (g *Generator) Generate(deviceID string, licenseType license.LicenseType, expiryDate time.Time, features []string) (string, error) {
	// 创建许可证对象
	lic := &license.License{
		DeviceID:    deviceID,
		ExpiryDate:  expiryDate,
		LicenseType: licenseType,
		Features:    features,
		CreatedAt:   time.Now(),
	}
	
	// 序列化为JSON
	jsonData, err := json.Marshal(lic)
	if err != nil {
		return "", err
	}
	
	// 使用AES加密
	encryptedData, err := crypto.EncryptAES(jsonData, g.aesKey)
	if err != nil {
		return "", err
	}
	
	// 使用RSA签名
	signature, err := crypto.SignData(encryptedData, g.privateKey)
	if err != nil {
		return "", err
	}
	
	// 组合数据：签名 + 密文
	licenseData := append(signature, encryptedData...)
	
	// Base64编码
	licenseKey := base64.StdEncoding.EncodeToString(licenseData)
	
	return licenseKey, nil
}

