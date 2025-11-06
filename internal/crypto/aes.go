// Package crypto 提供加密和解密功能
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// EncryptAES 使用AES-256-GCM加密数据
// 参数：
//   - plaintext: 明文数据
//   - key: 32字节的密钥（AES-256需要32字节）
//
// 返回值：
//   - []byte: 加密后的数据（包含nonce和密文）
//   - error: 加密过程中的错误
func EncryptAES(plaintext []byte, key []byte) ([]byte, error) {
	// 验证密钥长度
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes for AES-256")
	}

	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 生成nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// 加密
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptAES 使用AES-256-GCM解密数据
// 参数：
//   - ciphertext: 密文数据（包含nonce和密文）
//   - key: 32字节的密钥
//
// 返回值：
//   - []byte: 解密后的明文数据
//   - error: 解密过程中的错误
func DecryptAES(ciphertext []byte, key []byte) ([]byte, error) {
	// 验证密钥长度
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes for AES-256")
	}

	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 提取nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

