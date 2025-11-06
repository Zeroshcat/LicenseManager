// Package license 提供许可证生成和验证功能
package license

import "errors"

// 定义许可证相关的错误
var (
	// ErrInvalidLicense 表示许可证无效
	ErrInvalidLicense = errors.New("invalid license")

	// ErrExpiredLicense 表示许可证已过期
	ErrExpiredLicense = errors.New("license expired")

	// ErrDeviceMismatch 表示设备ID不匹配
	ErrDeviceMismatch = errors.New("device ID mismatch")

	// ErrNetworkError 表示网络验证失败
	ErrNetworkError = errors.New("network verification failed")

	// ErrLicenseNotFound 表示未找到许可证
	ErrLicenseNotFound = errors.New("license not found")

	// ErrInvalidKey 表示密钥无效
	ErrInvalidKey = errors.New("invalid key")
)

