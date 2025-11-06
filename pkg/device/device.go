// Package device 提供设备ID获取和硬件指纹识别功能
package device

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
)

// GetDeviceID 获取设备唯一ID（基于硬件指纹）
// 返回值：
//   - string: 设备ID（SHA256哈希值）
//   - error: 获取过程中的错误
func GetDeviceID() (string, error) {
	// 收集硬件信息
	info, err := collectHardwareInfo()
	if err != nil {
		return "", fmt.Errorf("failed to collect hardware info: %w", err)
	}
	
	// 生成哈希
	hash := sha256.Sum256([]byte(info))
	return hex.EncodeToString(hash[:]), nil
}

// collectHardwareInfo 收集硬件信息用于生成设备指纹
// 返回值：
//   - string: 硬件信息字符串
//   - error: 收集过程中的错误
func collectHardwareInfo() (string, error) {
	var info string
	
	// 获取主机名
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	info += hostname
	
	// 获取操作系统信息
	info += runtime.GOOS
	info += runtime.GOARCH
	
	// 获取用户信息
	info += os.Getenv("USER")
	info += os.Getenv("USERNAME")
	
	return info, nil
}


