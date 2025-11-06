package main

import (
	"fmt"
	"os"

	"github.com/Zeroshcat/LicenseManager/internal/crypto"
	licensegen "github.com/Zeroshcat/LicenseManager/internal/license"
	"github.com/Zeroshcat/LicenseManager/pkg/license"
)

func main() {
	fmt.Println("=== 测试密钥替换攻击场景 ===\n")

	// 场景：用户A 生成许可证
	fmt.Println("1. 用户A 生成密钥对和许可证...")
	
	// 用户A 的密钥
	privateKeyA, err := crypto.GenerateRSAKeyPair()
	if err != nil {
		fmt.Printf("❌ 生成密钥对失败: %v\n", err)
		os.Exit(1)
	}
	
	publicKeyA := &privateKeyA.PublicKey
	publicKeyAPEM, _ := crypto.EncodePublicKey(publicKeyA)
	
	aesKeyA := make([]byte, 32)
	for i := range aesKeyA {
		aesKeyA[i] = byte(i) // 模拟AES密钥
	}
	
	// 生成许可证
	generatorA := licensegen.NewGenerator(privateKeyA, aesKeyA)
	licenseKey, err := generatorA.Generate("device123", license.LicenseTypeOffline, 
		license.ParseExpiryDate("2026-12-31"), nil)
	if err != nil {
		fmt.Printf("❌ 生成许可证失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ 用户A 的许可证生成成功\n")

	// 场景：用户B 生成新密钥对（替换密钥）
	fmt.Println("2. 用户B 生成新密钥对（替换密钥）...")
	
	privateKeyB, err := crypto.GenerateRSAKeyPair()
	if err != nil {
		fmt.Printf("❌ 生成密钥对失败: %v\n", err)
		os.Exit(1)
	}
	
	publicKeyB := &privateKeyB.PublicKey
	publicKeyBPEM, _ := crypto.EncodePublicKey(publicKeyB)
	
	aesKeyB := make([]byte, 32)
	for i := range aesKeyB {
		aesKeyB[i] = byte(i + 100) // 不同的AES密钥
	}
	fmt.Println("✅ 用户B 的新密钥对生成成功\n")

	// 场景：用户B 用新密钥尝试验证用户A的许可证
	fmt.Println("3. 用户B 用新密钥尝试验证用户A的许可证...")
	
	verifierB, err := license.NewOfflineVerifier(publicKeyBPEM, aesKeyB)
	if err != nil {
		fmt.Printf("❌ 创建验证器失败: %v\n", err)
		os.Exit(1)
	}
	
	result, err := verifierB.Verify(licenseKey, "device123")
	if err != nil {
		fmt.Printf("❌ 验证失败: %v\n", err)
		fmt.Println("\n结论：用户B 无法用新密钥验证用户A的许可证！")
		fmt.Println("原因：RSA签名验证失败（公钥和私钥不匹配）")
	} else {
		fmt.Printf("✅ 验证成功: %+v\n", result)
		fmt.Println("\n⚠️  警告：这不应该发生！")
	}

	// 对比：用户A 用自己的密钥验证
	fmt.Println("\n4. 用户A 用自己的密钥验证自己的许可证...")
	
	verifierA, err := license.NewOfflineVerifier(publicKeyAPEM, aesKeyA)
	if err != nil {
		fmt.Printf("❌ 创建验证器失败: %v\n", err)
		os.Exit(1)
	}
	
	result, err = verifierA.Verify(licenseKey, "device123")
	if err != nil {
		fmt.Printf("❌ 验证失败: %v\n", err)
	} else {
		fmt.Printf("✅ 验证成功: %+v\n", result)
		fmt.Println("结论：只有用匹配的密钥才能验证许可证！")
	}
}

