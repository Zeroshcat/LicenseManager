package main

import (
	"embed"
	"fmt"
	"log"

	"github.com/Zeroshcat/LicenseManager/pkg/device"
	"github.com/Zeroshcat/LicenseManager/pkg/license"
)

//go:embed public_key.pem aes_key.bin
// 注意：编译前需要将 public_key.pem 和 aes_key.bin 复制到此目录
var embeddedKeys embed.FS

// loadEmbeddedKeys 从嵌入的文件系统加载密钥
func loadEmbeddedKeys() ([]byte, []byte, error) {
	// 读取嵌入的公钥
	publicKeyPEM, err := embeddedKeys.ReadFile("public_key.pem")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read embedded public key: %w", err)
	}

	// 读取嵌入的AES密钥
	aesKey, err := embeddedKeys.ReadFile("aes_key.bin")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read embedded AES key: %w", err)
	}

	return publicKeyPEM, aesKey, nil
}

func main() {
	// 获取设备ID
	deviceID, err := device.GetDeviceID()
	if err != nil {
		log.Fatalf("Failed to get device ID: %v", err)
	}

	// 从嵌入的文件系统加载密钥
	publicKeyPEM, aesKey, err := loadEmbeddedKeys()
	if err != nil {
		log.Fatalf("Failed to load embedded keys: %v", err)
	}

	// 创建离线验证器（使用嵌入的密钥）
	verifier, err := license.NewOfflineVerifier(publicKeyPEM, aesKey)
	if err != nil {
		log.Fatalf("Failed to create verifier: %v", err)
	}

	// 从文件加载许可证
	licenseKey, err := license.LoadLicenseFromFile("license.key")
	if err != nil {
		log.Fatalf("Failed to load license: %v", err)
	}

	// 验证许可证
	result, err := verifier.Verify(licenseKey, deviceID)
	if err != nil {
		log.Fatalf("License verification failed: %v", err)
	}

	if result.Valid && !result.Expired {
		fmt.Printf("✅ License is valid! Expires on: %s\n", result.ExpiryDate.Format("2006-01-02"))
	} else {
		fmt.Println("❌ License is invalid or expired")
	}
}

