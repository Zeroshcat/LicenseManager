package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Zeroshcat/LicenseManager/internal/crypto"
	licensegen "github.com/Zeroshcat/LicenseManager/internal/license"
	"github.com/Zeroshcat/LicenseManager/pkg/device"
	"github.com/Zeroshcat/LicenseManager/pkg/license"
)

func main() {
	fmt.Println("=== LicenseManager 测试程序 ===\n")

	// 1. 获取设备ID
	fmt.Println("1. 获取设备ID...")
	deviceID, err := device.GetDeviceID()
	if err != nil {
		fmt.Printf("❌ 获取设备ID失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ 设备ID: %s\n\n", deviceID)

	// 2. 检查密钥文件是否存在
	fmt.Println("2. 检查密钥文件...")
	if _, err := os.Stat("private_key.pem"); os.IsNotExist(err) {
		fmt.Println("❌ private_key.pem 不存在，请先运行 'licensemanager init'")
		os.Exit(1)
	}
	if _, err := os.Stat("public_key.pem"); os.IsNotExist(err) {
		fmt.Println("❌ public_key.pem 不存在，请先运行 'licensemanager init'")
		os.Exit(1)
	}
	if _, err := os.Stat("aes_key.bin"); os.IsNotExist(err) {
		fmt.Println("❌ aes_key.bin 不存在，请先运行 'licensemanager init'")
		os.Exit(1)
	}
	fmt.Println("✅ 密钥文件存在\n")

	// 3. 加载密钥
	fmt.Println("3. 加载密钥...")
	privateKeyPEM, err := os.ReadFile("private_key.pem")
	if err != nil {
		fmt.Printf("❌ 读取私钥失败: %v\n", err)
		os.Exit(1)
	}

	publicKeyPEM, err := os.ReadFile("public_key.pem")
	if err != nil {
		fmt.Printf("❌ 读取公钥失败: %v\n", err)
		os.Exit(1)
	}

	aesKey, err := os.ReadFile("aes_key.bin")
	if err != nil {
		fmt.Printf("❌ 读取AES密钥失败: %v\n", err)
		os.Exit(1)
	}

	privateKey, err := crypto.DecodePrivateKey(privateKeyPEM)
	if err != nil {
		fmt.Printf("❌ 解码私钥失败: %v\n", err)
		os.Exit(1)
	}

	if len(aesKey) != 32 {
		fmt.Printf("❌ AES密钥长度不正确: 期望32字节，实际%d字节\n", len(aesKey))
		os.Exit(1)
	}
	fmt.Println("✅ 密钥加载成功\n")

	// 4. 生成许可证
	fmt.Println("4. 生成许可证...")
	generator := licensegen.NewGenerator(privateKey, aesKey)
	expiryDate := time.Now().Add(365 * 24 * time.Hour) // 1年后过期

	licenseKey, err := generator.Generate(deviceID, license.LicenseTypeOffline, expiryDate, nil)
	if err != nil {
		fmt.Printf("❌ 生成许可证失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ 许可证生成成功\n")
	fmt.Printf("   许可证长度: %d 字符\n", len(licenseKey))
	fmt.Printf("   许可证前50字符: %s...\n\n", licenseKey[:min(50, len(licenseKey))])

	// 5. 保存到文件
	fmt.Println("5. 保存许可证到文件...")
	testLicenseFile := "test_license.key"
	if err := os.WriteFile(testLicenseFile, []byte(licenseKey), 0644); err != nil {
		fmt.Printf("❌ 保存许可证文件失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ 许可证已保存到: %s\n\n", testLicenseFile)

	// 6. 从文件读取并验证
	fmt.Println("6. 从文件读取并验证许可证...")
	loadedLicenseKey, err := license.LoadLicenseFromFile(testLicenseFile)
	if err != nil {
		fmt.Printf("❌ 读取许可证文件失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ 许可证读取成功\n")
	fmt.Printf("   读取的许可证长度: %d 字符\n", len(loadedLicenseKey))
	fmt.Printf("   读取的许可证前50字符: %s...\n\n", loadedLicenseKey[:min(50, len(loadedLicenseKey))])

	// 7. 创建验证器并验证
	fmt.Println("7. 创建验证器...")
	verifier, err := license.NewOfflineVerifier(publicKeyPEM, aesKey)
	if err != nil {
		fmt.Printf("❌ 创建验证器失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ 验证器创建成功\n")

	// 8. 验证许可证
	fmt.Println("8. 验证许可证...")
	result, err := verifier.Verify(loadedLicenseKey, deviceID)
	if err != nil {
		fmt.Printf("❌ 验证失败: %v\n", err)
		os.Exit(1)
	}

	if result.Valid && !result.Expired {
		fmt.Println("✅ 许可证验证成功！")
		fmt.Printf("   设备ID: %s\n", result.DeviceID)
		fmt.Printf("   许可证类型: %s\n", result.LicenseType)
		fmt.Printf("   到期时间: %s\n", result.ExpiryDate.Format("2006-01-02 15:04:05"))
		fmt.Printf("   消息: %s\n", result.Message)
	} else {
		fmt.Println("❌ 许可证验证失败")
		fmt.Printf("   有效: %v\n", result.Valid)
		fmt.Printf("   过期: %v\n", result.Expired)
		fmt.Printf("   消息: %s\n", result.Message)
		os.Exit(1)
	}

	// 9. 清理测试文件
	fmt.Println("\n9. 清理测试文件...")
	if err := os.Remove(testLicenseFile); err != nil {
		fmt.Printf("⚠️  删除测试文件失败: %v\n", err)
	} else {
		fmt.Println("✅ 测试文件已清理")
	}

	fmt.Println("\n=== 所有测试通过！ ===")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

