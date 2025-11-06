// Package database 提供数据库操作功能
package database

import (
	"time"

	"gorm.io/gorm"
)

// LicenseRecord 许可证记录
type LicenseRecord struct {
	ID          int64          `gorm:"primaryKey;autoIncrement"` // 主键ID
	DeviceID    string         `gorm:"not null;index"`           // 设备ID
	LicenseKey  string         `gorm:"not null"`                 // 许可证密钥
	LicenseType string         `gorm:"not null"`                 // 许可证类型
	ExpiryDate  time.Time      `gorm:"not null"`                 // 到期时间
	CreatedAt   time.Time      // 创建时间
	UpdatedAt   time.Time      // 更新时间
	DeletedAt   gorm.DeletedAt `gorm:"index"` // 软删除
}

// TableName 指定表名
func (LicenseRecord) TableName() string {
	return "licenses"
}

// DeviceRecord 设备记录
type DeviceRecord struct {
	ID           int64          `gorm:"primaryKey;autoIncrement"` // 主键ID
	DeviceID     string         `gorm:"uniqueIndex;not null"`     // 设备ID（唯一）
	DeviceName   string         // 设备名称
	AppID        string         // 应用ID
	LicenseID    int64          // 关联的许可证ID
	Status       string         `gorm:"default:active;index"` // 状态（active, expired, revoked）
	RegisteredAt time.Time      `gorm:"not null"`             // 注册时间
	LastSeen     time.Time      `gorm:"not null"`             // 最后访问时间
	DeletedAt    gorm.DeletedAt `gorm:"index"`                // 软删除
}

// TableName 指定表名
func (DeviceRecord) TableName() string {
	return "devices"
}

// KeyRecord 密钥记录
type KeyRecord struct {
	ID        int64          `gorm:"primaryKey;autoIncrement"` // 主键ID
	KeyType   string         `gorm:"not null"`                 // 密钥类型（aes, rsa_private, rsa_public）
	KeyData   []byte         `gorm:"not null"`                 // 密钥数据
	CreatedAt time.Time      `gorm:"not null"`                 // 创建时间
	DeletedAt gorm.DeletedAt `gorm:"index"`                    // 软删除
}

// TableName 指定表名
func (KeyRecord) TableName() string {
	return "keys"
}
