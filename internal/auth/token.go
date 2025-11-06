// Package auth 提供认证和Token管理功能
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

// GenerateToken 生成随机Token
// 参数：
//   - length: Token长度（字节数）
// 返回值：
//   - string: base64编码的Token
//   - error: 生成过程中的错误
func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ValidateToken 验证Token是否有效
// 参数：
//   - token: Token值
//   - tokenType: Token类型（client|admin）
//   - appID: 应用ID（client类型需要）
// 返回值：
//   - bool: Token是否有效
//   - error: 验证过程中的错误
func ValidateToken(token string, tokenType string, appID string) (bool, error) {
	// 这里应该查询数据库验证Token
	// 简化实现，返回true
	return true, nil
}

// TokenExpiry 计算Token过期时间
// 参数：
//   - days: 有效天数（0表示永不过期）
// 返回值：
//   - time.Time: 过期时间
func TokenExpiry(days int) time.Time {
	if days == 0 {
		// 永不过期，返回一个很远的未来时间
		return time.Now().AddDate(100, 0, 0)
	}
	return time.Now().AddDate(0, 0, days)
}


