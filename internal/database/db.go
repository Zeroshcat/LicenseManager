// Package database 提供数据库操作功能
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite" // 纯Go SQLite驱动，无需CGO
)

// DB 数据库连接
type DB struct {
	db *gorm.DB
}

// NewDB 创建数据库连接
// 参数：
//   - dbPath: 数据库文件路径
//
// 返回值：
//   - *DB: 数据库实例
//   - error: 连接过程中的错误
func NewDB(dbPath string) (*DB, error) {
	// 转换为绝对路径，避免路径问题
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		// 如果无法获取绝对路径，使用原始路径
		absPath = dbPath
	}

	// 确保数据库文件所在目录存在
	dir := filepath.Dir(absPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory %s: %w", dir, err)
		}
	}

	// 检查目录权限
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to stat directory %s: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", dir)
	}

	// 检查是否可写
	testFile := filepath.Join(dir, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return nil, fmt.Errorf("directory is not writable: %s (%w)", dir, err)
	}
	os.Remove(testFile) // 清理测试文件

	// 清理可能存在的SQLite锁文件和临时文件
	lockFiles := []string{
		absPath + "-shm",     // WAL共享内存文件
		absPath + "-wal",     // WAL日志文件
		absPath + "-journal", // 回滚日志文件
	}
	for _, lockFile := range lockFiles {
		if _, err := os.Stat(lockFile); err == nil {
			// 锁文件存在，尝试删除（可能来自之前的异常退出）
			os.Remove(lockFile)
		}
	}

	// 先使用database/sql和modernc.org/sqlite打开连接
	// 这样可以确保使用纯Go驱动，而不是需要CGO的go-sqlite3
	sqlDB, err := sql.Open("sqlite", absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database (path: %s): %w", absPath, err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(1) // SQLite不支持并发写入
	sqlDB.SetMaxIdleConns(1)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 使用GORM包装已打开的database/sql连接
	// 使用sqlite.Dialector的Conn字段传入已创建的连接
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 静默模式，减少日志输出
	})
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to open gorm database (path: %s): %w", absPath, err)
	}

	// 设置SQLite参数
	pragmas := []string{
		"PRAGMA busy_timeout = 10000",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA foreign_keys = ON",
	}

	for _, pragma := range pragmas {
		if err := db.Exec(pragma).Error; err != nil {
			// 如果某些PRAGMA失败，记录但不阻止初始化
			continue
		}
	}

	database := &DB{db: db}

	// 自动迁移表结构
	if err := database.autoMigrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate tables: %w", err)
	}

	return database, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	sqlDB, err := db.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// autoMigrate 自动迁移数据库表结构
// 返回值：
//   - error: 迁移过程中的错误
func (db *DB) autoMigrate() error {
	// 自动创建/更新表结构
	return db.db.AutoMigrate(
		&LicenseRecord{},
		&DeviceRecord{},
		&KeyRecord{},
		&TokenRecord{},
	)
}

// SaveLicense 保存许可证记录
// 参数：
//   - record: 许可证记录
//
// 返回值：
//   - int64: 插入的记录ID
//   - error: 保存过程中的错误
func (db *DB) SaveLicense(record *LicenseRecord) (int64, error) {
	now := time.Now()
	record.CreatedAt = now
	record.UpdatedAt = now

	if err := db.db.Create(record).Error; err != nil {
		return 0, err
	}

	return record.ID, nil
}

// GetLicenseByDeviceID 根据设备ID获取许可证
// 参数：
//   - deviceID: 设备ID
//
// 返回值：
//   - *LicenseRecord: 许可证记录
//   - error: 查询过程中的错误
func (db *DB) GetLicenseByDeviceID(deviceID string) (*LicenseRecord, error) {
	var record LicenseRecord
	if err := db.db.Where("device_id = ?", deviceID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("license not found for device: %s", deviceID)
		}
		return nil, err
	}

	return &record, nil
}

// SaveDevice 保存设备记录
// 参数：
//   - record: 设备记录
//
// 返回值：
//   - int64: 插入的记录ID
//   - error: 保存过程中的错误
func (db *DB) SaveDevice(record *DeviceRecord) (int64, error) {
	now := time.Now()
	record.RegisteredAt = now
	record.LastSeen = now

	if err := db.db.Create(record).Error; err != nil {
		return 0, err
	}

	return record.ID, nil
}

// GetDeviceByID 根据设备ID获取设备记录
// 参数：
//   - deviceID: 设备ID
//
// 返回值：
//   - *DeviceRecord: 设备记录
//   - error: 查询过程中的错误
func (db *DB) GetDeviceByID(deviceID string) (*DeviceRecord, error) {
	var record DeviceRecord
	if err := db.db.Where("device_id = ?", deviceID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("device not found: %s", deviceID)
		}
		return nil, err
	}

	return &record, nil
}

// ListDevices 列出所有设备
// 参数：
//   - limit: 限制数量
//   - offset: 偏移量
//
// 返回值：
//   - []*DeviceRecord: 设备记录列表
//   - error: 查询过程中的错误
func (db *DB) ListDevices(limit, offset int) ([]*DeviceRecord, error) {
	var devices []*DeviceRecord
	if err := db.db.Order("registered_at DESC").Limit(limit).Offset(offset).Find(&devices).Error; err != nil {
		return nil, err
	}

	return devices, nil
}

// GetStats 获取统计数据
// 返回值：
//   - map[string]int64: 统计数据
//   - error: 查询过程中的错误
func (db *DB) GetStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	// 总设备数
	var totalDevices int64
	if err := db.db.Model(&DeviceRecord{}).Count(&totalDevices).Error; err != nil {
		return nil, err
	}
	stats["total_devices"] = totalDevices

	// 活跃设备数
	var activeDevices int64
	if err := db.db.Model(&DeviceRecord{}).Where("status = ?", "active").Count(&activeDevices).Error; err != nil {
		return nil, err
	}
	stats["active_devices"] = activeDevices

	// 总许可证数
	var totalLicenses int64
	if err := db.db.Model(&LicenseRecord{}).Count(&totalLicenses).Error; err != nil {
		return nil, err
	}
	stats["total_licenses"] = totalLicenses

	// 已过期许可证数
	now := time.Now()
	var expiredLicenses int64
	if err := db.db.Model(&LicenseRecord{}).Where("expiry_date < ?", now).Count(&expiredLicenses).Error; err != nil {
		return nil, err
	}
	stats["expired_licenses"] = expiredLicenses

	return stats, nil
}

// ListLicenses 列出所有许可证
// 参数：
//   - limit: 限制数量
//   - offset: 偏移量
//
// 返回值：
//   - []*LicenseRecord: 许可证记录列表
//   - error: 查询过程中的错误
func (db *DB) ListLicenses(limit, offset int) ([]*LicenseRecord, error) {
	var licenses []*LicenseRecord
	if err := db.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&licenses).Error; err != nil {
		return nil, err
	}

	return licenses, nil
}

// GetLicenseByID 根据ID获取许可证
// 参数：
//   - id: 许可证ID
//
// 返回值：
//   - *LicenseRecord: 许可证记录
//   - error: 查询过程中的错误
func (db *DB) GetLicenseByID(id int64) (*LicenseRecord, error) {
	var record LicenseRecord
	if err := db.db.Where("id = ?", id).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("license not found: %d", id)
		}
		return nil, err
	}

	return &record, nil
}

// DeleteLicense 删除许可证（软删除）
// 参数：
//   - id: 许可证ID
//
// 返回值：
//   - error: 删除过程中的错误
func (db *DB) DeleteLicense(id int64) error {
	return db.db.Delete(&LicenseRecord{}, id).Error
}

// UpdateDeviceStatus 更新设备状态
// 参数：
//   - deviceID: 设备ID
//   - status: 新状态
//
// 返回值：
//   - error: 更新过程中的错误
func (db *DB) UpdateDeviceStatus(deviceID, status string) error {
	return db.db.Model(&DeviceRecord{}).Where("device_id = ?", deviceID).Update("status", status).Error
}

// GetDeviceCount 获取设备总数
// 返回值：
//   - int64: 设备总数
//   - error: 查询过程中的错误
func (db *DB) GetDeviceCount() (int64, error) {
	var count int64
	if err := db.db.Model(&DeviceRecord{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
