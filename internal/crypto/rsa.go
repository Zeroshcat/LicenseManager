// Package crypto 提供加密和解密功能
package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// GenerateRSAKeyPair 生成RSA-4096密钥对
// 返回值：
//   - *rsa.PrivateKey: 私钥
//   - *rsa.PublicKey: 公钥
//   - error: 生成过程中的错误
func GenerateRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// 生成4096位RSA密钥对
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}
	
	return privateKey, &privateKey.PublicKey, nil
}

// SignData 使用RSA私钥签名数据
// 参数：
//   - data: 要签名的数据
//   - privateKey: RSA私钥
// 返回值：
//   - []byte: 签名数据
//   - error: 签名过程中的错误
func SignData(data []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	// 计算哈希
	hash := sha256.Sum256(data)
	
	// 使用PSS签名
	signature, err := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, hash[:], nil)
	if err != nil {
		return nil, err
	}
	
	return signature, nil
}

// VerifySignature 使用RSA公钥验证签名
// 参数：
//   - data: 原始数据
//   - signature: 签名数据
//   - publicKey: RSA公钥
// 返回值：
//   - bool: 签名是否有效
//   - error: 验证过程中的错误
func VerifySignature(data []byte, signature []byte, publicKey *rsa.PublicKey) (bool, error) {
	// 计算哈希
	hash := sha256.Sum256(data)
	
	// 验证签名
	err := rsa.VerifyPSS(publicKey, crypto.SHA256, hash[:], signature, nil)
	if err != nil {
		return false, err
	}
	
	return true, nil
}

// EncodePrivateKey 将RSA私钥编码为PEM格式
// 参数：
//   - privateKey: RSA私钥
// 返回值：
//   - []byte: PEM编码的私钥
func EncodePrivateKey(privateKey *rsa.PrivateKey) []byte {
	keyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}
	return pem.EncodeToMemory(block)
}

// EncodePublicKey 将RSA公钥编码为PEM格式
// 参数：
//   - publicKey: RSA公钥
// 返回值：
//   - []byte: PEM编码的公钥
func EncodePublicKey(publicKey *rsa.PublicKey) []byte {
	keyBytes, _ := x509.MarshalPKIXPublicKey(publicKey)
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: keyBytes,
	}
	return pem.EncodeToMemory(block)
}

// DecodePrivateKey 从PEM格式解码RSA私钥
// 参数：
//   - pemData: PEM编码的私钥数据
// 返回值：
//   - *rsa.PrivateKey: RSA私钥
//   - error: 解码过程中的错误
func DecodePrivateKey(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	
	return privateKey, nil
}

// DecodePublicKey 从PEM格式解码RSA公钥
// 参数：
//   - pemData: PEM编码的公钥数据
// 返回值：
//   - *rsa.PublicKey: RSA公钥
//   - error: 解码过程中的错误
func DecodePublicKey(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	
	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	
	return rsaPublicKey, nil
}

