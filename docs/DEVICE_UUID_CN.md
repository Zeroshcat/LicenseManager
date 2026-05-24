# 设备UUID方案说明

[English](DEVICE_UUID.md) | 简体中文

## 概述

LicenseManager 使用标准UUID格式作为设备唯一标识符。设备UUID由稳定硬件序列号派生，不直接使用系统安装ID。

企业授权场景下，设备绑定应当与物理设备一致，而不是与某一次系统安装一致。因此当前策略是：能获取可信硬件序列号才生成设备UUID，获取不到就返回错误。

## UUID格式

设备UUID采用标准UUID格式：`XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX`

示例：`F6235A40-C9E2-5681-B236-ED9C4C15E58D`

## 生成策略

### 硬件序列号优先

设备UUID只基于主板或整机硬件序列号生成：

- **Linux**：优先读取 DMI 主板序列号 `/sys/class/dmi/id/board_serial`，必要时调用 `dmidecode -s baseboard-serial-number`
- **macOS**：使用 `system_profiler SPHardwareDataType` 获取整机序列号
- **Windows**：优先使用 PowerShell/CIM 获取 `Win32_BaseBoard.SerialNumber`，必要时使用 `wmic baseboard get SerialNumber`

获取到的硬件序列号会先规范化，再通过哈希派生为UUID。许可证中保存的是派生后的设备UUID，不保存原始硬件序列号。

### 不使用的标识

以下标识不会作为设备UUID来源：

- `/etc/machine-id`、`/var/lib/dbus/machine-id`
- 主机名
- MAC地址
- 操作系统名称、架构、安装时间等系统安装态信息
- 明显无效的硬件占位值，例如 `To Be Filled By O.E.M.`、`Not Specified`、`Default string`、全0或全F序列号

这些值容易因为重装系统、网络环境变化、虚拟化配置或厂商占位数据而改变，不符合企业授权绑定预期。

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

deviceUUID, err := device.GetDeviceID()
if err != nil {
    log.Fatalf("Failed to get device UUID: %v", err)
}

fmt.Printf("Device UUID: %s\n", deviceUUID)
```

## 行为边界

1. **重装系统**：只要主板/整机硬件序列号不变，设备UUID应保持不变，许可证仍可用。
2. **更换主板或整机**：设备UUID会变化，原许可证不应继续可用。
3. **无法读取硬件序列号**：命令直接失败，需要修复系统权限、DMI数据或厂商占位序列号问题后再签发许可证。
4. **既有许可证迁移**：如果旧版本曾使用系统UUID生成许可证，升级到硬件序列号策略后设备UUID会变化，需要重新签发许可证或提供一次性迁移流程。

## 故障排查

### 问题：无法获取设备UUID

**解决方案**：

1. 确认设备能读取硬件序列号：
   - Linux: `cat /sys/class/dmi/id/board_serial` 或 `sudo dmidecode -s baseboard-serial-number`
   - macOS: `system_profiler SPHardwareDataType`
   - Windows: `Get-CimInstance Win32_BaseBoard | Select-Object SerialNumber`
2. 如果返回 `To Be Filled By O.E.M.`、`Not Specified` 等占位值，需要由设备厂商、BIOS/固件或运维流程修正。
3. 如果读取命令需要更高权限，应在部署文档中明确运行权限要求。

### 问题：同一硬件不同系统下UUID不一致

**说明**：

- Linux 和 Windows 通常都能读取主板序列号，理论上同一硬件应一致。
- macOS 使用整机序列号，和其他平台的主板序列来源不同。
- 多系统共存或虚拟化场景应在目标运行环境中获取UUID，并以该环境的输出作为签发依据。

## 设计取舍

- 当前实现选择 **fail closed**：宁可无法生成设备UUID，也不静默降级到不稳定标识。
- 这会减少某些设备的开箱成功率，但授权边界更清晰，能避免系统重装后误判为新设备。
- 如果业务需要支持无硬件序列号的设备，应显式设计人工审核、服务器侧设备注册或企业资产编号绑定流程，而不是在本地自动兜底。
