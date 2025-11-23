# 设备UUID方案说明

## 概述

LicenseManager 使用标准UUID格式作为设备唯一标识符，参考了 [xinjiayu/LicenseManager](https://github.com/xinjiayu/LicenseManager) 项目的实现方案。

## UUID格式

设备UUID采用标准UUID格式：`XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX`

示例：`F6235A40-C9E2-5681-B236-ED9C4C15E58D`

## 生成策略

### 优先级1：系统UUID（推荐）

优先使用操作系统提供的机器UUID，这是最稳定的方案：

#### Linux
- **主要来源**：`/etc/machine-id`
- **备用来源**：`/var/lib/dbus/machine-id`
- **特点**：系统级唯一标识，不会因硬件更换而变化

#### macOS
- **主要来源**：`ioreg` 命令获取 `IOPlatformUUID`
- **备用来源**：`system_profiler SPHardwareDataType` 获取 Hardware UUID
- **特点**：硬件级别的唯一标识，与硬件绑定

#### Windows
- **主要来源**：`wmic csproduct get UUID` 获取 ComputerSystem UUID
- **备用来源**：PowerShell `Get-WmiObject Win32_ComputerSystemProduct`
- **特点**：主板级别的唯一标识，相对稳定

### 优先级2：硬件信息生成UUID（后备方案）

如果系统UUID不可用，将基于以下硬件信息生成UUID格式标识符：

1. **MAC地址**：优先使用第一个非虚拟网络接口的MAC地址
2. **主机名**：系统主机名
3. **CPU信息**：
   - Linux: `/proc/cpuinfo`
   - macOS: `sysctl machdep.cpu.brand_string`
   - Windows: `wmic cpu get ProcessorId`
4. **磁盘序列号**：
   - Linux: `lsblk`
   - macOS: `diskutil info`
   - Windows: `wmic diskdrive get SerialNumber`
5. **主板序列号**：
   - Linux: `dmidecode`
   - macOS: `system_profiler`
   - Windows: `wmic baseboard get SerialNumber`
6. **操作系统信息**：`GOOS` 和 `GOARCH`

这些信息通过MD5哈希后转换为标准UUID格式。

## 使用方法

### 命令行获取UUID

```bash
# 获取当前设备的UUID
./licensemanager uuid
# 或
./licensemanager checkuuid

# 输出示例
F6235A40-C9E2-5681-B236-ED9C4C15E58D
```

### Go程序中使用

```go
import "github.com/Zeroshcat/LicenseManager/pkg/device"

// 获取设备UUID
deviceUUID, err := device.GetDeviceID()
if err != nil {
    log.Fatalf("Failed to get device UUID: %v", err)
}

// deviceUUID 格式: F6235A40-C9E2-5681-B236-ED9C4C15E58D
fmt.Printf("Device UUID: %s\n", deviceUUID)
```

## 优势

1. **稳定性**：系统UUID不会因用户修改主机名、更换网络接口等操作而变化
2. **唯一性**：UUID格式确保全局唯一性
3. **兼容性**：与参考项目格式一致，便于迁移和集成
4. **可读性**：标准UUID格式易于识别和使用
5. **跨平台**：支持 Linux、macOS、Windows 三大主流平台

## 注意事项

1. **系统UUID优先**：系统UUID是最稳定的方案，建议确保系统UUID可用
2. **虚拟环境**：在虚拟机中，系统UUID通常是稳定的，但MAC地址可能变化
3. **容器环境**：容器中的系统UUID可能来自宿主机，需要注意隔离性
4. **硬件更换**：如果更换了主板（Windows）或主要硬件（macOS），系统UUID可能会变化

## 故障排查

### 问题：无法获取系统UUID

**解决方案**：
1. 检查系统权限（某些系统UUID文件需要root权限）
2. 确认系统UUID文件存在：
   - Linux: `cat /etc/machine-id`
   - macOS: `ioreg -rd1 -c IOPlatformExpertDevice | grep IOPlatformUUID`
   - Windows: `wmic csproduct get UUID`

### 问题：UUID格式不正确

**解决方案**：
- 代码会自动将非标准格式转换为标准UUID格式
- 如果仍有问题，检查系统UUID文件内容

### 问题：UUID在不同平台不一致

**说明**：
- 这是正常现象，不同平台使用不同的UUID来源
- 同一设备在不同平台上的UUID可能不同
- 建议在同一平台上生成和验证许可证

## 参考

- [xinjiayu/LicenseManager](https://github.com/xinjiayu/LicenseManager) - 参考项目
- [RFC 4122](https://tools.ietf.org/html/rfc4122) - UUID标准规范

