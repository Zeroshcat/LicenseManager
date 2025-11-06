// Package database 提供数据库操作功能
package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// TokenRecord Token记录
type TokenRecord struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`        // 主键ID
	Token     string         `gorm:"uniqueIndex;not null" json:"token"`         // Token值
	TokenType string         `gorm:"not null" json:"token_type"`                // Token类型（client|admin）
	AppID     string         `json:"app_id"`                                    // 应用ID（client类型需要）
	CreatedAt time.Time      `gorm:"not null" json:"created_at"`                // 创建时间
	ExpiresAt *time.Time     `json:"expires_at"`                                // 过期时间（可选，使用指针支持NULL）
	Revoked   bool           `gorm:"default:false" json:"revoked"`               // 是否已撤销
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`                            // 软删除（不序列化）
}

// TableName 指定表名
func (TokenRecord) TableName() string {
	return "tokens"
}

// SaveToken 保存Token记录
// 参数：
//   - record: Token记录
//
// 返回值：
//   - int64: 插入的记录ID
//   - error: 保存过程中的错误
func (db *DB) SaveToken(record *TokenRecord) (int64, error) {
	record.CreatedAt = time.Now()
	record.Revoked = false

	if err := db.db.Create(record).Error; err != nil {
		return 0, err
	}

	return record.ID, nil
}

// GetToken 根据Token值获取Token记录
// 参数：
//   - token: Token值
//
// 返回值：
//   - *TokenRecord: Token记录
//   - error: 查询过程中的错误
func (db *DB) GetToken(token string) (*TokenRecord, error) {
	var record TokenRecord
	if err := db.db.Where("token = ?", token).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("token not found")
		}
		return nil, err
	}

	return &record, nil
}

// RevokeToken 撤销Token
// 参数：
//   - token: Token值
//
// 返回值：
//   - error: 撤销过程中的错误
func (db *DB) RevokeToken(token string) error {
	return db.db.Model(&TokenRecord{}).Where("token = ?", token).Update("revoked", true).Error
}

// ListTokens 列出所有Token
// 参数：
//   - limit: 限制数量
//   - offset: 偏移量
//
// 返回值：
//   - []*TokenRecord: Token记录列表
//   - error: 查询过程中的错误
func (db *DB) ListTokens(limit, offset int) ([]*TokenRecord, error) {
	var tokens []*TokenRecord
	if err := db.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&tokens).Error; err != nil {
		return nil, err
	}

	return tokens, nil
}
