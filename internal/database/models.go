// Package database 提供数据库操作功能
package database

import (
	"time"

	"gorm.io/gorm"
)

// LicenseRecord 许可证记录
type LicenseRecord struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"` // 主键ID
	DeviceID    string         `gorm:"not null;index" json:"device_id"`    // 设备ID
	LicenseKey  string         `gorm:"not null" json:"license_key"`        // 许可证密钥
	LicenseType string         `gorm:"not null" json:"license_type"`       // 许可证类型
	ExpiryDate  time.Time      `gorm:"not null" json:"expiry_date"`        // 到期时间
	CreatedAt   time.Time      `json:"created_at"`                         // 创建时间
	UpdatedAt   time.Time      `json:"updated_at"`                         // 更新时间
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`                     // 软删除（不序列化）
}

// TableName 指定表名
func (LicenseRecord) TableName() string {
	return "licenses"
}

// DeviceRecord 设备记录
type DeviceRecord struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`    // 主键ID
	DeviceID     string         `gorm:"uniqueIndex;not null" json:"device_id"` // 设备ID（唯一）
	DeviceName   string         `json:"device_name"`                           // 设备名称
	AppID        string         `json:"app_id"`                                // 应用ID
	LicenseID    int64          `json:"license_id"`                            // 关联的许可证ID
	Status       string         `gorm:"default:active;index" json:"status"`    // 状态（active, expired, revoked）
	RegisteredAt time.Time      `gorm:"not null" json:"registered_at"`         // 注册时间
	LastSeen     time.Time      `gorm:"not null" json:"last_seen"`             // 最后访问时间
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`                        // 软删除（不序列化）
}

// TableName 指定表名
func (DeviceRecord) TableName() string {
	return "devices"
}

// KeyRecord 密钥记录
type KeyRecord struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"` // 主键ID
	KeyType   string         `gorm:"not null" json:"key_type"`           // 密钥类型（aes, rsa_private, rsa_public）
	KeyData   []byte         `gorm:"not null" json:"-"`                  // 密钥数据（不序列化，安全考虑）
	CreatedAt time.Time      `gorm:"not null" json:"created_at"`         // 创建时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`                     // 软删除（不序列化）
}

// TableName 指定表名
func (KeyRecord) TableName() string {
	return "keys"
}
