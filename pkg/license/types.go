// Package license 提供许可证生成和验证功能
package license

import "time"

// LicenseType 许可证类型
type LicenseType string

const (
	// LicenseTypeOffline 离线许可证
	LicenseTypeOffline LicenseType = "offline"
	
	// LicenseTypeOnline 网络许可证
	LicenseTypeOnline LicenseType = "online"
	
	// LicenseTypeDual 双重验证许可证
	LicenseTypeDual LicenseType = "dual"
)

// License 许可证结构
type License struct {
	DeviceID    string      // 设备ID
	ExpiryDate  time.Time   // 到期时间
	LicenseType LicenseType // 许可证类型
	Features    []string    // 功能列表
	CreatedAt   time.Time   // 创建时间
}

// VerifyResult 验证结果
type VerifyResult struct {
	Valid        bool      // 是否有效
	Expired      bool      // 是否过期
	ExpiryDate   time.Time // 到期时间
	DeviceID     string    // 设备ID
	LicenseType  string    // 许可证类型
	OfflineValid bool      // 离线验证结果（仅双重验证）
	OnlineValid  bool      // 网络验证结果（仅双重验证和网络验证）
	Message      string    // 验证消息
}

// Verifier 验证器接口
type Verifier interface {
	// Verify 验证许可证
	// 参数根据验证器类型不同而不同
	Verify(...interface{}) (*VerifyResult, error)
}


