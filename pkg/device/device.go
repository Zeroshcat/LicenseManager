// Package device 提供设备ID获取和硬件指纹识别功能
// 参考: https://github.com/xinjiayu/LicenseManager
package device

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

// GetDeviceID 获取设备唯一UUID（参考 xinjiayu/LicenseManager 方案）
// 优先使用系统提供的机器UUID，如果获取不到则基于硬件信息生成UUID格式的标识符
// 返回值：
//   - string: 设备UUID（标准UUID格式：XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX）
//   - error: 获取过程中的错误
func GetDeviceID() (string, error) {
	// 1. 优先尝试获取系统机器UUID
	uuid, err := getSystemUUID()
	if err == nil && uuid != "" {
		// 如果已经是标准UUID格式，直接返回
		if isValidUUID(uuid) {
			return formatUUID(uuid), nil
		}
		// 如果不是标准格式，转换为UUID格式
		return convertToUUID(uuid), nil
	}

	// 2. 如果系统UUID不可用，基于硬件信息生成UUID
	return generateHardwareUUID()
}

// getSystemUUID 获取系统提供的机器UUID
// Linux: /etc/machine-id
// macOS: Hardware UUID (IOPlatformUUID)
// Windows: ComputerSystem UUID
func getSystemUUID() (string, error) {
	switch runtime.GOOS {
	case "linux":
		// Linux: 读取 /etc/machine-id
		data, err := os.ReadFile("/etc/machine-id")
		if err == nil {
			uuid := strings.TrimSpace(string(data))
			if uuid != "" {
				return uuid, nil
			}
		}
		// 尝试 /var/lib/dbus/machine-id
		data, err = os.ReadFile("/var/lib/dbus/machine-id")
		if err == nil {
			uuid := strings.TrimSpace(string(data))
			if uuid != "" {
				return uuid, nil
			}
		}

	case "darwin":
		// macOS: 使用 ioreg 获取 IOPlatformUUID（更可靠的方法）
		cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
		output, err := cmd.Output()
		if err == nil {
			// 查找 IOPlatformUUID
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "IOPlatformUUID") {
					// 提取UUID值
					re := regexp.MustCompile(`"IOPlatformUUID" = "([^"]+)"`)
					matches := re.FindStringSubmatch(line)
					if len(matches) > 1 {
						return matches[1], nil
					}
				}
			}
		}
		// 备用方案: system_profiler
		cmd = exec.Command("system_profiler", "SPHardwareDataType")
		output, err = cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "Hardware UUID") {
					parts := strings.Split(line, ":")
					if len(parts) > 1 {
						uuid := strings.TrimSpace(parts[1])
						if uuid != "" {
							return uuid, nil
						}
					}
				}
			}
		}

	case "windows":
		// Windows: 使用 wmic 获取 ComputerSystem UUID
		cmd := exec.Command("wmic", "csproduct", "get", "UUID")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && line != "UUID" && len(line) > 10 {
					// 清理可能的额外空格
					uuid := strings.Fields(line)[0]
					return uuid, nil
				}
			}
		}
		// 备用方案: 使用 PowerShell 获取
		cmd = exec.Command("powershell", "-Command", "(Get-WmiObject Win32_ComputerSystemProduct).UUID")
		output, err = cmd.Output()
		if err == nil {
			uuid := strings.TrimSpace(string(output))
			if uuid != "" && len(uuid) > 10 {
				return uuid, nil
			}
		}
	}

	return "", fmt.Errorf("system UUID not available on %s", runtime.GOOS)
}

// generateHardwareUUID 基于硬件信息生成UUID格式的设备标识符
func generateHardwareUUID() (string, error) {
	var parts []string

	// 1. 获取MAC地址（优先使用）
	macAddrs, err := getMACAddresses()
	if err == nil && len(macAddrs) > 0 {
		for _, mac := range macAddrs {
			if mac != "" && !isVirtualMAC(mac) {
				parts = append(parts, mac)
				break
			}
		}
	}

	// 2. 获取主机名
	hostname, err := os.Hostname()
	if err == nil && hostname != "" {
		parts = append(parts, hostname)
	}

	// 3. 获取CPU信息
	cpuID, err := getCPUInfo()
	if err == nil && cpuID != "" {
		parts = append(parts, cpuID)
	}

	// 4. 获取磁盘序列号
	diskID, err := getDiskSerial()
	if err == nil && diskID != "" {
		parts = append(parts, diskID)
	}

	// 5. 获取主板序列号
	boardID, err := getBoardSerial()
	if err == nil && boardID != "" {
		parts = append(parts, boardID)
	}

	// 6. 操作系统和架构信息
	parts = append(parts, runtime.GOOS, runtime.GOARCH)

	if len(parts) == 0 {
		return "", fmt.Errorf("failed to collect hardware information")
	}

	// 组合信息并生成MD5哈希，然后转换为UUID格式
	info := strings.Join(parts, "|")
	return convertToUUID(info), nil
}

// convertToUUID 将字符串转换为UUID格式
// 使用MD5哈希生成16字节，然后格式化为标准UUID格式
func convertToUUID(input string) string {
	hash := md5.Sum([]byte(input))
	hexStr := hex.EncodeToString(hash[:])

	// 格式化为标准UUID格式: XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hexStr[0:8],
		hexStr[8:12],
		hexStr[12:16],
		hexStr[16:20],
		hexStr[20:32])
}

// formatUUID 格式化UUID字符串为标准格式
func formatUUID(uuid string) string {
	// 移除所有连字符和空格，转换为大写
	cleaned := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(uuid, "-", ""), " ", ""))

	if len(cleaned) != 32 {
		// 如果不是32个字符，使用MD5哈希
		hash := md5.Sum([]byte(uuid))
		cleaned = hex.EncodeToString(hash[:])
	}

	// 格式化为标准UUID格式
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		cleaned[0:8],
		cleaned[8:12],
		cleaned[12:16],
		cleaned[16:20],
		cleaned[20:32])
}

// isValidUUID 检查字符串是否为有效的UUID格式
func isValidUUID(uuid string) bool {
	// UUID格式: XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
	uuidPattern := regexp.MustCompile(`^[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}$`)
	return uuidPattern.MatchString(uuid)
}

// getMACAddresses 获取所有网络接口的MAC地址
func getMACAddresses() ([]string, error) {
	var macs []string
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	
	for _, iface := range interfaces {
		// 跳过回环接口和未激活的接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		
		mac := iface.HardwareAddr.String()
		if mac != "" {
			macs = append(macs, mac)
		}
	}
	
	return macs, nil
}

// isVirtualMAC 判断是否为虚拟MAC地址
func isVirtualMAC(mac string) bool {
	// 常见的虚拟MAC地址前缀
	virtualPrefixes := []string{
		"00:00:00:00:00:00",
		"00:05:69", // VMware
		"00:0c:29", // VMware
		"00:50:56", // VMware
		"00:1c:14", // VMware
		"08:00:27", // VirtualBox
		"0a:00:27", // VirtualBox
		"52:54:00", // QEMU/KVM
	}
	
	macLower := strings.ToLower(mac)
	for _, prefix := range virtualPrefixes {
		if strings.HasPrefix(macLower, strings.ToLower(prefix)) {
			return true
		}
	}
	return false
}


// getCPUInfo 获取CPU信息
func getCPUInfo() (string, error) {
	switch runtime.GOOS {
	case "linux":
		// 读取 /proc/cpuinfo
		data, err := os.ReadFile("/proc/cpuinfo")
		if err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "Serial") || strings.HasPrefix(line, "processor") {
					parts := strings.Split(line, ":")
					if len(parts) > 1 {
						value := strings.TrimSpace(parts[1])
						if value != "" && value != "0" {
							return value, nil
						}
					}
				}
			}
		}
		
	case "darwin":
		// macOS: 使用 sysctl 获取 CPU 品牌
		cmd := exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output)), nil
		}
		
	case "windows":
		// Windows: 使用 wmic 获取 CPU ID
		cmd := exec.Command("wmic", "cpu", "get", "ProcessorId")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && line != "ProcessorId" && len(line) > 5 {
					return line, nil
				}
			}
		}
	}
	
	return "", fmt.Errorf("CPU info not available on %s", runtime.GOOS)
}

// getDiskSerial 获取磁盘序列号
func getDiskSerial() (string, error) {
	switch runtime.GOOS {
	case "linux":
		// 尝试读取 /dev/disk/by-id/ 或使用 lsblk
		cmd := exec.Command("lsblk", "-no", "SERIAL", "-d")
		output, err := cmd.Output()
		if err == nil {
			serial := strings.TrimSpace(string(output))
			if serial != "" {
				return serial, nil
			}
		}
		
	case "darwin":
		// macOS: 使用 diskutil
		cmd := exec.Command("diskutil", "info", "/")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "Volume UUID") {
					parts := strings.Split(line, ":")
					if len(parts) > 1 {
						return strings.TrimSpace(parts[1]), nil
					}
				}
			}
		}
		
	case "windows":
		// Windows: 使用 wmic 获取磁盘序列号
		cmd := exec.Command("wmic", "diskdrive", "get", "SerialNumber")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && line != "SerialNumber" && len(line) > 5 {
					return line, nil
				}
			}
		}
	}
	
	return "", fmt.Errorf("disk serial not available on %s", runtime.GOOS)
}

// getBoardSerial 获取主板序列号
func getBoardSerial() (string, error) {
	switch runtime.GOOS {
	case "linux":
		// 读取 DMI 信息
		cmd := exec.Command("dmidecode", "-s", "baseboard-serial-number")
		output, err := cmd.Output()
		if err == nil {
			serial := strings.TrimSpace(string(output))
			if serial != "" && serial != "Not Specified" {
				return serial, nil
			}
		}
		
	case "darwin":
		// macOS: 使用 system_profiler
		cmd := exec.Command("system_profiler", "SPHardwareDataType")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "Serial Number") {
					parts := strings.Split(line, ":")
					if len(parts) > 1 {
						return strings.TrimSpace(parts[1]), nil
					}
				}
			}
		}
		
	case "windows":
		// Windows: 使用 wmic 获取主板序列号
		cmd := exec.Command("wmic", "baseboard", "get", "SerialNumber")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && line != "SerialNumber" && len(line) > 5 {
					return line, nil
				}
			}
		}
	}
	
	return "", fmt.Errorf("board serial not available on %s", runtime.GOOS)
}


