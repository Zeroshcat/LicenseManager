// Package device 提供设备ID获取和硬件指纹识别功能
package device

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

// GetDeviceID 获取设备唯一UUID。
//
// 企业授权绑定只接受稳定硬件标识，当前策略为主板/整机硬件序列号优先。
// 如果无法获取可信硬件序列号，直接返回错误，不降级到 machine-id、主机名、MAC 等可变态信息。
// 返回值：
//   - string: 设备UUID（标准UUID格式：XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX）
//   - error: 获取过程中的错误
func GetDeviceID() (string, error) {
	serial, err := getBoardSerial()
	if err != nil {
		return "", fmt.Errorf("failed to get stable hardware serial: %w", err)
	}

	return hardwareIDToUUID("mainboard", serial), nil
}

// getBoardSerial 获取主板或整机硬件序列号。
func getBoardSerial() (string, error) {
	switch runtime.GOOS {
	case "linux":
		// 优先读取 sysfs DMI 信息；读取不到时再尝试 dmidecode。
		if serial, err := readFirstUsableFile("/sys/class/dmi/id/board_serial"); err == nil {
			return serial, nil
		}

		cmd := exec.Command("dmidecode", "-s", "baseboard-serial-number")
		output, err := cmd.Output()
		if err == nil {
			if serial := firstUsableOutputLine(output); serial != "" {
				return serial, nil
			}
		}

	case "darwin":
		cmd := exec.Command("system_profiler", "SPHardwareDataType")
		output, err := cmd.Output()
		if err == nil {
			for _, line := range strings.Split(string(output), "\n") {
				if strings.Contains(line, "Serial Number") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) != 2 {
						continue
					}
					serial := canonicalHardwareValue(parts[1])
					if isUsableHardwareID(serial) {
						return serial, nil
					}
				}
			}
		}

	case "windows":
		commands := []*exec.Cmd{
			exec.Command("powershell", "-NoProfile", "-Command", "(Get-CimInstance Win32_BaseBoard).SerialNumber"),
			exec.Command("powershell", "-NoProfile", "-Command", "(Get-WmiObject Win32_BaseBoard).SerialNumber"),
			exec.Command("wmic", "baseboard", "get", "SerialNumber"),
		}

		for _, cmd := range commands {
			output, err := cmd.Output()
			if err != nil {
				continue
			}
			if serial := firstUsableOutputLine(output, "SerialNumber"); serial != "" {
				return serial, nil
			}
		}
	}

	return "", fmt.Errorf("mainboard serial not available on %s", runtime.GOOS)
}

func hardwareIDToUUID(source, value string) string {
	input := canonicalHardwareValue(source) + ":" + canonicalHardwareValue(value)
	hash := sha256.Sum256([]byte(input))
	bytes := hash[:16]

	// 生成内部确定性UUID，不暴露原始硬件序列号。
	bytes[6] = (bytes[6] & 0x0f) | 0x50
	bytes[8] = (bytes[8] & 0x3f) | 0x80

	hexStr := strings.ToUpper(hex.EncodeToString(bytes))
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hexStr[0:8],
		hexStr[8:12],
		hexStr[12:16],
		hexStr[16:20],
		hexStr[20:32])
}

// isValidUUID 检查字符串是否为有效的UUID格式。
func isValidUUID(uuid string) bool {
	uuidPattern := regexp.MustCompile(`^[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}$`)
	return uuidPattern.MatchString(uuid)
}

func readFirstUsableFile(paths ...string) (string, error) {
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if value := firstUsableOutputLine(data); value != "" {
			return value, nil
		}
	}

	return "", fmt.Errorf("no usable hardware ID file found")
}

func firstUsableOutputLine(output []byte, headers ...string) string {
	headerSet := make(map[string]struct{}, len(headers))
	for _, header := range headers {
		headerSet[canonicalHardwareValue(header)] = struct{}{}
	}

	for _, line := range strings.Split(string(output), "\n") {
		value := canonicalHardwareValue(line)
		if value == "" {
			continue
		}
		if _, ok := headerSet[value]; ok {
			continue
		}
		if isUsableHardwareID(value) {
			return value
		}
	}

	return ""
}

func canonicalHardwareValue(value string) string {
	value = strings.ReplaceAll(value, "\x00", "")
	value = strings.TrimPrefix(value, "\ufeff")
	return strings.ToUpper(strings.Join(strings.Fields(strings.TrimSpace(value)), " "))
}

func isUsableHardwareID(value string) bool {
	value = canonicalHardwareValue(value)
	if value == "" {
		return false
	}

	normalized := strings.NewReplacer(" ", "", "-", "", "_", "", ".", "", ":", "", "/", "").Replace(value)
	if normalized == "" || allSameChar(normalized) {
		return false
	}

	invalidValues := map[string]struct{}{
		"0":                  {},
		"NA":                 {},
		"N/A":                {},
		"NONE":               {},
		"NULL":               {},
		"OEM":                {},
		"UNKNOWN":            {},
		"INVALID":            {},
		"NOTAVAILABLE":       {},
		"NOTAPPLICABLE":      {},
		"NOTSPECIFIED":       {},
		"SERIALNUMBER":       {},
		"SYSTEMSERIALNUMBER": {},
		"DEFAULT":            {},
		"DEFAULTSTRING":      {},
		"TOBEFILLEDBYOEM":    {},
	}
	if _, ok := invalidValues[normalized]; ok {
		return false
	}

	invalidFragments := []string{
		"TOBEFILLEDBY",
		"NOTSPECIFIED",
		"NOTAVAILABLE",
		"SYSTEMSERIALNUMBER",
		"DEFAULTSTRING",
	}
	for _, fragment := range invalidFragments {
		if strings.Contains(normalized, fragment) {
			return false
		}
	}

	return true
}

func allSameChar(value string) bool {
	if value == "" {
		return true
	}

	first := value[0]
	for i := 1; i < len(value); i++ {
		if value[i] != first {
			return false
		}
	}
	return true
}
