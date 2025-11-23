# LicenseManager - 统一许可证管理工具

一个基于 Golang 开发的企业级许可证管理系统，支持离线授权、网络授权和双重验证模式。

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 特性亮点

- 🔐 **多种授权模式**：离线、在线、双重验证三种模式满足不同场景需求
- 🛡️ **安全加密**：AES-256-GCM + RSA-4096 双重加密保护
- 🖥️ **Web 管理界面**：现代化的 Web 界面，支持在线生成和管理许可证
- 🔑 **密码保护**：管理界面支持密码保护，确保安全访问
- 🗄️ **SQLite 数据库**：轻量级数据库，无需额外配置

## 功能特性

### 核心功能
- ✅ **多种授权模式**
  - 离线授权：基于本地密钥文件的授权验证
  - 网络授权：基于服务器验证的在线授权
  - 双重验证：同时需要离线密钥和网络验证才能授权
  
- ✅ **设备绑定**
  - 基于系统UUID的设备唯一标识
  - 优先使用系统提供的机器UUID（Linux/macOS/Windows）
  - 标准UUID格式：`XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX`
  - 支持多设备管理
  - 设备授权状态追踪

- ✅ **到期管理**
  - 灵活的许可证有效期设置
  - 自动到期检测
  - 到期提醒功能

- ✅ **安全加密**
  - 使用 AES-256-GCM 加密算法
  - RSA-4096 密钥对用于签名验证
  - 安全的密钥存储和传输

### 管理功能
- ✅ **后台管理界面**
  - Web 管理界面（可选）
  - 密码保护，安全访问
  - 密钥创建和管理
  - 在线许可证生成和分发
  - 设备授权状态查看
  - Token 管理

- ✅ **数据库**
  - SQLite 轻量级数据库
  - 完整的许可证和设备记录

### 工具特性
- ✅ **统一命令行工具**
  - 所有功能集成在单一 CLI 工具中
  - 清晰的命令结构

- ✅ **多种输出格式**
  - 文本格式（人类可读）
  - JSON 格式（程序化处理）

## 项目结构

```
LicenseManager/
├── README.md
├── go.mod
├── go.sum
├── cmd/
│   ├── licensemanager/      # 主命令行工具
│   └── server/              # 授权服务器（可选）
├── internal/
│   ├── crypto/              # 加密相关功能
│   ├── license/             # 许可证生成和验证
│   ├── device/              # 设备绑定管理
│   ├── database/            # 数据库操作
│   ├── server/              # 网络授权服务器
│   └── admin/               # 后台管理
├── pkg/
│   ├── device/              # 设备UUID获取（参考 xinjiayu/LicenseManager）
│   ├── license/             # 许可证验证器
│   └── output/              # 输出格式化
├── web/                     # Web 管理界面（可选）
├── examples/                # 示例代码
│   └── embedded_keys/      # 嵌入密钥示例
└── config/                  # 配置文件示例
```

## 安装

### 前置要求
- Go 1.21 或更高版本
- SQLite3（使用纯 Go 实现，无需 CGO）

### 构建

```bash
# 克隆仓库
git clone https://github.com/Zeroshcat/LicenseManager.git
cd LicenseManager

# 构建主工具
go build -o licensemanager ./cmd/licensemanager

# 构建授权服务器（可选）
go build -o license-server ./cmd/server
```

## 快速开始

### 1. 初始化系统

首次使用需要初始化数据库和生成密钥：

```bash
# 初始化数据库和生成主密钥
./licensemanager init
```

```bash
# 指定数据库路径
./licensemanager init --db /data/license.db
```

初始化完成后，会在当前目录生成以下文件：
- `license.db` - SQLite 数据库文件
- `private_key.pem` - RSA 私钥（用于签名）
- `public_key.pem` - RSA 公钥（用于验证）
- `aes_key.bin` - AES 加密密钥

⚠️ **重要**：请妥善保管这些密钥文件，丢失后无法恢复！

### 2. 生成许可证

#### 命令行生成

**步骤1：获取设备UUID**
```bash
# 获取目标设备的UUID
./licensemanager uuid
# 输出: F6235A40-C9E2-5681-B236-ED9C4C15E58D
```

**步骤2：生成许可证**
```bash
# 生成离线许可证（使用设备UUID）
./licensemanager generate --type offline --device-id F6235A40-C9E2-5681-B236-ED9C4C15E58D --expiry 2024-12-31

# 生成网络许可证
./licensemanager generate --type online --device-id F6235A40-C9E2-5681-B236-ED9C4C15E58D --expiry 2024-12-31

# 生成双重验证许可证
./licensemanager generate --type dual --device-id F6235A40-C9E2-5681-B236-ED9C4C15E58D --expiry 2024-12-31

# 保存到文件
./licensemanager generate --type offline --device-id F6235A40-C9E2-5681-B236-ED9C4C15E58D --expiry 2024-12-31 --output license.key

# 如果不指定 --device-id，将自动获取当前设备的UUID
./licensemanager generate --type offline --expiry 2024-12-31 --output license.key
```

#### Web 界面生成

1. 启动管理服务器：`./licensemanager admin serve --passwd your_password`
2. 访问 Web 界面：`http://localhost:8080`
3. 登录后进入"许可证管理"标签页
4. 点击"生成新许可证"按钮
5. 填写设备UUID（可通过 `./licensemanager uuid` 获取）、选择许可证类型、设置到期日期
6. 点击"生成"即可创建许可证，生成的许可证密钥会自动显示并可复制
7. 生成成功后可以点击"下载 license.key"按钮直接下载许可证文件
8. 在许可证列表中，每个许可证都有"下载"按钮，可以随时下载

### 3. 验证许可证

```bash
# 验证离线许可证（自动获取设备ID）
./licensemanager verify --license-file license.key

# 验证离线许可证（指定设备UUID）
./licensemanager verify --license-file license.key --device-id F6235A40-C9E2-5681-B236-ED9C4C15E58D

# 验证网络许可证
./licensemanager verify --online --device-id F6235A40-C9E2-5681-B236-ED9C4C15E58D --api-url http://localhost:8080

# 验证双重验证许可证
./licensemanager verify --dual --license-file license.key --device-id F6235A40-C9E2-5681-B236-ED9C4C15E58D --api-url http://localhost:8080
```

**注意**：验证时会自动清理许可证文件中的换行符和空格，确保 Base64 编码正确解析。

### 4. 获取设备UUID

```bash
# 获取当前设备的UUID（用于生成许可证）
./licensemanager uuid
# 或
./licensemanager checkuuid

# 输出示例: F6235A40-C9E2-5681-B236-ED9C4C15E58D
```

**设备UUID说明**：
- **Linux**: 使用 `/etc/machine-id` 或 `/var/lib/dbus/machine-id`
- **macOS**: 使用 `ioreg` 获取 `IOPlatformUUID` 或 `system_profiler` 获取 Hardware UUID
- **Windows**: 使用 `wmic csproduct get UUID` 或 PowerShell 获取 ComputerSystem UUID
- 如果系统UUID不可用，将基于硬件信息（MAC地址、CPU、磁盘等）生成UUID格式标识符

### 5. 设备管理

```bash
# 列出所有设备
./licensemanager device list

# 查看当前设备UUID
./licensemanager device show

# 查看指定设备详情
./licensemanager device show <device-uuid>

# 绑定设备
./licensemanager device bind <device-uuid>
```

### 后台管理

```bash
# 启动管理服务器（必须提供密码）
./licensemanager admin serve --passwd your_password --port 8080

# 指定其他参数
./licensemanager admin serve --passwd your_password --host 0.0.0.0 --port 8080 --db license.db

# 访问 Web 界面
# http://localhost:8080
# 首次访问需要输入管理密码登录
```

#### Web 管理界面功能

- **密码保护**：所有管理功能都需要密码验证
- **统计概览**：查看总设备数、活跃设备、许可证数量等统计信息
- **设备管理**：查看所有注册设备，包括设备UUID、名称、状态、注册时间等
- **许可证管理**：
  - 查看所有许可证列表（ID、设备UUID、类型、到期时间、创建时间）
  - 在线生成新许可证（支持离线/在线/双重验证三种类型）
  - 生成后可直接下载 `license.key` 文件
  - 许可证列表中每个许可证都支持快捷下载
  - 删除许可证
- **Token 管理**：查看和管理 API Token，支持撤销操作

### 7. API Token 管理

```bash
# 创建客户端 Token
./licensemanager admin token create --type client --app-id your_app_id

# 创建管理员 Token
./licensemanager admin token create --type admin

# 创建带过期时间的 Token
./licensemanager admin token create --type client --app-id your_app_id --expires 2024-12-31
```

### 8. 输出格式

```bash
# 文本输出（默认）
./licensemanager device list

# JSON 输出
./licensemanager device list --format json
```

## Go 程序集成使用方法

LicenseManager 提供了 Go 包，方便其他 Go 程序通过 `import` 方式集成许可证验证功能。

### 安装

```bash
go get github.com/Zeroshcat/LicenseManager/pkg/license
go get github.com/Zeroshcat/LicenseManager/pkg/device
```

### 使用方式

#### 1. 离线验证（完全本地，无需网络）

离线验证完全在本地进行，不需要任何网络连接，适合内网环境或离线场景。

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/Zeroshcat/LicenseManager/pkg/license"
    "github.com/Zeroshcat/LicenseManager/pkg/device"
)

func main() {
    // 获取设备UUID（标准UUID格式，如：F6235A40-C9E2-5681-B236-ED9C4C15E58D）
    deviceUUID, err := device.GetDeviceID()
    if err != nil {
        log.Fatalf("Failed to get device UUID: %v", err)
    }
    
    // 创建离线验证器（不需要API地址）
    verifier := license.NewOfflineVerifier()
    
    // 从文件读取许可证密钥
    licenseKey, err := license.LoadLicenseFromFile("license.key")
    if err != nil {
        log.Fatalf("Failed to load license: %v", err)
    }
    
    // 验证许可证（完全本地验证，无需网络）
    result, err := verifier.Verify(licenseKey, deviceUUID)
    if err != nil {
        log.Fatalf("License verification failed: %v", err)
    }
    
    if result.Valid && !result.Expired {
        fmt.Printf("License is valid! Expires on: %s\n", result.ExpiryDate)
    } else {
        fmt.Println("License is invalid or expired")
    }
}
```

#### 2. 网络验证（需要预设API地址）

网络验证需要连接到授权服务器，适合需要实时验证和控制的场景。

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/Zeroshcat/LicenseManager/pkg/license"
    "github.com/Zeroshcat/LicenseManager/pkg/device"
)

func main() {
    // 获取设备ID
    deviceID, err := device.GetDeviceID()
    if err != nil {
        log.Fatalf("Failed to get device ID: %v", err)
    }
    
    // 创建网络验证器（需要预设API地址）
    // API地址在初始化时设置，后续验证都使用此地址
    verifier := license.NewOnlineVerifier(&license.OnlineConfig{
        APIURL: "https://license.yourcompany.com/api/v1", // 预设API地址
        AppID:  "your_application_id",
        Timeout: 10, // 超时时间（秒）
    })
    
    // 验证许可证（通过网络验证）
    result, err := verifier.Verify(deviceUUID)
    if err != nil {
        log.Fatalf("License verification failed: %v", err)
    }
    
    if result.Valid && !result.Expired {
        fmt.Printf("License is valid! Expires on: %s\n", result.ExpiryDate)
    } else {
        fmt.Println("License is invalid or expired")
    }
}
```

#### 3. 双重验证（离线 + 网络，两者都必须通过）

双重验证需要同时满足离线验证和网络验证，提供最高级别的安全性。

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/Zeroshcat/LicenseManager/pkg/license"
    "github.com/Zeroshcat/LicenseManager/pkg/device"
)

func main() {
    // 获取设备ID
    deviceID, err := device.GetDeviceID()
    if err != nil {
        log.Fatalf("Failed to get device ID: %v", err)
    }
    
    // 加载离线许可证
    licenseKey, err := license.LoadLicenseFromFile("license.key")
    if err != nil {
        log.Fatalf("Failed to load license: %v", err)
    }
    
    // 创建双重验证器
    // 需要同时提供离线许可证和网络API地址
    verifier := license.NewDualVerifier(&license.DualConfig{
        APIURL: "https://license.yourcompany.com/api/v1", // 预设API地址
        AppID:  "your_application_id",
        Timeout: 10,
    })
    
    // 验证许可证（需要同时通过离线验证和网络验证）
    result, err := verifier.Verify(licenseKey, deviceUUID)
    if err != nil {
        log.Fatalf("License verification failed: %v", err)
    }
    
    if result.Valid && !result.Expired {
        fmt.Printf("License is valid! Expires on: %s\n", result.ExpiryDate)
        fmt.Printf("Offline verification: %v\n", result.OfflineValid)
        fmt.Printf("Online verification: %v\n", result.OnlineValid)
    } else {
        fmt.Println("License is invalid or expired")
        if !result.OfflineValid {
            fmt.Println("Offline verification failed")
        }
        if !result.OnlineValid {
            fmt.Println("Online verification failed")
        }
    }
}
```

### 完整集成示例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/Zeroshcat/LicenseManager/pkg/license"
    "github.com/Zeroshcat/LicenseManager/pkg/device"
)

type App struct {
    verifier  license.Verifier
    deviceUUID string
}

func NewApp(licenseType string, apiURL string) (*App, error) {
    deviceUUID, err := device.GetDeviceID()
    if err != nil {
        return nil, fmt.Errorf("failed to get device UUID: %w", err)
    }
    
    var verifier license.Verifier
    
    switch licenseType {
    case "offline":
        // 离线验证：完全本地，无需网络
        verifier = license.NewOfflineVerifier()
        
    case "online":
        // 网络验证：需要预设API地址
        verifier = license.NewOnlineVerifier(&license.OnlineConfig{
            APIURL: apiURL,
            AppID:  os.Getenv("APP_ID"),
            Timeout: 10,
        })
        
    case "dual":
        // 双重验证：需要离线许可证和网络API地址
        verifier = license.NewDualVerifier(&license.DualConfig{
            APIURL: apiURL,
            AppID:  os.Getenv("APP_ID"),
            Timeout: 10,
        })
        
    default:
        return nil, fmt.Errorf("unknown license type: %s", licenseType)
    }
    
    return &App{
        verifier:   verifier,
        deviceUUID: deviceUUID,
    }, nil
}

func (app *App) CheckLicense() error {
    var result *license.VerifyResult
    var err error
    
    // 根据验证器类型调用不同的验证方法
    switch v := app.verifier.(type) {
    case *license.OfflineVerifier:
        licenseKey, err := license.LoadLicenseFromFile("license.key")
        if err != nil {
            return fmt.Errorf("failed to load license: %w", err)
        }
        result, err = v.Verify(licenseKey, app.deviceUUID)
        
    case *license.OnlineVerifier:
        result, err = v.Verify(app.deviceUUID)
        
    case *license.DualVerifier:
        licenseKey, err := license.LoadLicenseFromFile("license.key")
        if err != nil {
            return fmt.Errorf("failed to load license: %w", err)
        }
        result, err = v.Verify(licenseKey, app.deviceUUID)
    }
    
    if err != nil {
        return fmt.Errorf("verification failed: %w", err)
    }
    
    if !result.Valid || result.Expired {
        return fmt.Errorf("license is invalid or expired")
    }
    
    fmt.Printf("License verified successfully. Expires: %s\n", result.ExpiryDate)
    return nil
}

func main() {
    // 从环境变量或配置文件读取设置
    licenseType := os.Getenv("LICENSE_TYPE") // "offline", "online", "dual"
    apiURL := os.Getenv("API_URL")            // 仅网络验证和双重验证需要
    
    app, err := NewApp(licenseType, apiURL)
    if err != nil {
        log.Fatalf("Failed to initialize app: %v", err)
    }
    
    // 启动时验证
    if err := app.CheckLicense(); err != nil {
        log.Fatalf("License check failed: %v", err)
    }
    
    // 定期验证（可选）
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if err := app.CheckLicense(); err != nil {
                log.Printf("License check failed: %v", err)
                // 根据业务需求决定是否退出
            }
        }
    }
}
```

### API 配置说明

#### 离线验证配置

离线验证不需要任何配置，完全本地验证：

```go
verifier := license.NewOfflineVerifier()
// 无需API地址，无需网络连接
```

#### 网络验证配置

网络验证需要预设API地址：

```go
config := &license.OnlineConfig{
    APIURL:  "https://license.yourcompany.com/api/v1", // 必须：预设API地址
    AppID:   "your_application_id",                    // 必须：应用ID
    Timeout: 10,                                       // 可选：超时时间（秒）
    Retries: 3,                                        // 可选：重试次数
}
verifier := license.NewOnlineVerifier(config)
```

#### 双重验证配置

双重验证需要同时配置离线许可证和网络API地址：

```go
config := &license.DualConfig{
    APIURL:  "https://license.yourcompany.com/api/v1", // 必须：预设API地址
    AppID:   "your_application_id",                      // 必须：应用ID
    Timeout: 10,                                         // 可选：超时时间（秒）
    // 离线许可证通过文件加载，不在配置中
}
verifier := license.NewDualVerifier(config)
```

### 错误处理

```go
result, err := verifier.Verify(...)
if err != nil {
    switch err {
    case license.ErrInvalidLicense:
        fmt.Println("许可证无效")
    case license.ErrExpiredLicense:
        fmt.Println("许可证已过期")
    case license.ErrDeviceMismatch:
        fmt.Println("设备UUID不匹配")
    case license.ErrNetworkError:
        fmt.Println("网络验证失败（仅网络验证和双重验证）")
    default:
        fmt.Printf("验证错误: %v\n", err)
    }
}
```

### 验证结果结构

```go
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
```

## 开发规范

### 代码注释规范

- **所有公共函数和类型必须有注释**：使用 Go 标准注释格式
- **复杂逻辑必须添加行内注释**：解释关键步骤和业务逻辑
- **包级别注释**：每个包文件开头必须有包说明
- **函数注释格式**：
  ```go
  // FunctionName 函数功能描述
  // 参数说明：
  //   - param1: 参数1的说明
  //   - param2: 参数2的说明
  // 返回值说明：
  //   - 返回值1: 说明
  //   - error: 错误信息
  func FunctionName(param1 string, param2 int) (string, error) {
      // 实现代码
  }
  ```

### 函数设计规范

- **函数职责单一**：每个函数只做一件事，保持简洁
- **避免深层嵌套**：使用早期返回（early return）减少嵌套层级
- **函数长度控制**：单个函数不超过 50 行，复杂逻辑拆分为多个小函数
- **参数数量限制**：函数参数不超过 5 个，超过时使用结构体
- **错误处理**：所有可能失败的操作都要返回 error，不要忽略错误

### 代码示例

```go
// Package license 提供许可证生成和验证功能
package license

// VerifyResult 表示许可证验证结果
type VerifyResult struct {
    Valid      bool      // 是否有效
    Expired    bool      // 是否过期
    ExpiryDate time.Time // 到期时间
}

// Verify 验证许可证是否有效
// 参数：
//   - licenseKey: 许可证密钥
//   - deviceID: 设备ID
// 返回值：
//   - *VerifyResult: 验证结果
//   - error: 验证过程中的错误
func Verify(licenseKey string, deviceID string) (*VerifyResult, error) {
    // 参数验证
    if licenseKey == "" {
        return nil, ErrInvalidLicense
    }
    
    // 解析许可证
    license, err := parseLicense(licenseKey)
    if err != nil {
        return nil, err
    }
    
    // 检查设备ID
    if license.DeviceID != deviceID {
        return nil, ErrDeviceMismatch
    }
    
    // 检查是否过期
    if time.Now().After(license.ExpiryDate) {
        return &VerifyResult{
            Valid:      false,
            Expired:    true,
            ExpiryDate: license.ExpiryDate,
        }, nil
    }
    
    return &VerifyResult{
        Valid:      true,
        Expired:    false,
        ExpiryDate: license.ExpiryDate,
    }, nil
}

// parseLicense 解析许可证密钥
// 这是一个私有函数，用于内部实现
func parseLicense(key string) (*License, error) {
    // 实现解析逻辑
}
```

## 安全说明

### 密钥安全

#### 密钥文件说明

**离线验证需要的文件：**
- `public_key.pem`（公钥）：**可以公开分发** ✅
  - 用于验证许可证签名
  - 即使泄露也无法伪造许可证
  - 可以嵌入到客户端程序中
  
- `aes_key.bin`（AES密钥）：**需要保护** ⚠️
  - 用于解密许可证内容
  - 如果泄露，攻击者可以查看许可证内容，但无法伪造新许可证
  - 建议使用代码混淆或嵌入到程序中
  - **重要**：即使 AES 密钥泄露，攻击者也无法破解其他用户的许可证（因为需要对应的私钥才能生成有效签名）
  - **注意**：如果用户替换了密钥文件（运行 `init`），原许可证将无法验证，但这不是破解，用户只能生成自己的许可证

- `private_key.pem`（私钥）：**绝对不能泄露** 🚨
  - 只存在于服务器端
  - 用于生成和签名许可证
  - 如果泄露，攻击者可以伪造任意许可证

#### 安全建议
- 所有密钥文件应妥善保管，不要提交到版本控制系统
- 生产环境建议使用环境变量或密钥管理服务
- 定期轮换主密钥
- 网络授权建议使用 HTTPS
- 对于高安全要求，考虑使用在线验证模式

#### 嵌入密钥到代码中（推荐用于客户端）

对于客户端程序，可以将公钥和 AES 密钥嵌入到代码中，并使用代码混淆工具保护：

**方法1：使用 Go embed 指令（推荐）**

```go
package main

import (
    "embed"
    "github.com/Zeroshcat/LicenseManager/pkg/license"
)

//go:embed public_key.pem aes_key.bin
var embeddedKeys embed.FS

func main() {
    // 从嵌入的文件系统读取密钥
    publicKeyPEM, _ := embeddedKeys.ReadFile("public_key.pem")
    aesKey, _ := embeddedKeys.ReadFile("aes_key.bin")
    
    // 创建验证器
    verifier, _ := license.NewOfflineVerifier(publicKeyPEM, aesKey)
    // ... 使用验证器
}
```

**编译时使用 garble 混淆（强烈推荐）：**

```bash
# 安装 garble
go install mvdan.cc/garble@latest

# 使用 garble 混淆编译
garble build -o app main.go
```

**garble 的作用：**
- 混淆变量名和函数名
- 混淆字符串常量（包括嵌入的密钥）
- 增加逆向工程难度
- 减小二进制文件大小

**完整示例：**

1. **准备密钥文件**（编译前必须完成）：
   ```bash
   # Linux/macOS
   cp public_key.pem examples/embedded_keys/
   cp aes_key.bin examples/embedded_keys/
   
   # Windows
   copy public_key.pem examples\embedded_keys\
   copy aes_key.bin examples\embedded_keys\
   ```

2. **编译程序**：
   ```bash
   cd examples/embedded_keys
   go build -o app main.go
   ```

3. **使用 garble 混淆编译**（推荐）：
   ```bash
   # 安装 garble
   go install mvdan.cc/garble@latest
   
   # 混淆编译
   garble build -o app main.go
   ```

详细示例代码和更多方法请参考 [examples/embedded_keys/](examples/embedded_keys/)

#### 详细安全分析
更多安全分析请参考 [docs/SECURITY.md](docs/SECURITY.md)

#### 设备UUID方案
详细的设备UUID生成方案和使用说明，请参考 [docs/DEVICE_UUID.md](docs/DEVICE_UUID.md)

#### 设备UUID方案
详细的设备UUID生成方案和使用说明，请参考 [docs/DEVICE_UUID.md](docs/DEVICE_UUID.md)

### 后台管理安全
- **密码保护**：启动管理服务器时必须设置强密码（`--passwd` 参数）
- **会话管理**：登录后使用 Cookie 保持会话，默认 24 小时有效
- **访问控制**：所有管理 API 都需要密码验证
- **生产环境建议**：
  - 使用 HTTPS 访问管理界面
  - 定期更换管理密码
  - 限制管理服务器的访问 IP
  - 不要在公网直接暴露管理界面

## 常见问题

### Q: 如何重置管理密码？

A: 管理密码在启动服务器时通过 `--passwd` 参数设置，每次启动都需要提供。如果需要更改密码，只需使用新的密码重新启动服务器即可。

### Q: 密钥文件丢失了怎么办？

A: 密钥文件一旦丢失无法恢复。如果丢失了密钥文件：
1. 无法生成新的许可证
2. 无法验证已生成的许可证
3. 需要重新运行 `init` 命令生成新的密钥对
4. 注意：重新生成密钥后，之前使用旧密钥生成的许可证将无法验证

### Q: 如何在生产环境部署？

A: 生产环境部署建议：
1. 使用 HTTPS 访问管理界面
2. 将密钥文件存储在安全的位置（不要放在代码仓库中）
3. 使用环境变量或密钥管理服务管理敏感信息
4. 限制管理服务器的访问 IP（使用防火墙或反向代理）
5. 定期备份数据库和密钥文件
6. 定期轮换密钥和密码

### Q: 支持哪些数据库？

A: 目前使用 SQLite 数据库，使用纯 Go 实现（modernc.org/sqlite），无需 CGO，可以在任何平台运行。

### Q: 如何获取设备UUID？

A: 设备UUID基于系统UUID或硬件信息生成，标准UUID格式。可以通过以下方式获取：
```bash
# 使用命令行工具（推荐）
./licensemanager uuid
# 或
./licensemanager checkuuid

# 查看设备信息（包含UUID）
./licensemanager device show

# 或在 Go 程序中
import "github.com/Zeroshcat/LicenseManager/pkg/device"
deviceUUID, err := device.GetDeviceID()
// 返回格式: F6235A40-C9E2-5681-B236-ED9C4C15E58D
```

**UUID生成策略**：
1. **优先使用系统UUID**（最稳定）：
   - Linux: `/etc/machine-id`
   - macOS: `IOPlatformUUID`
   - Windows: `ComputerSystem UUID`
2. **后备方案**：如果系统UUID不可用，基于硬件信息（MAC地址、CPU、磁盘、主板等）生成UUID格式标识符

### Q: 许可证可以转移吗？

A: 许可证与设备UUID绑定，不能直接转移。如果需要在新设备上使用，需要：
1. 删除旧设备的许可证
2. 获取新设备的UUID：`./licensemanager uuid`
3. 为新设备生成新的许可证：`./licensemanager generate --device-id <new-uuid> ...`

### Q: 如何测试许可证功能？

A: 可以使用项目根目录下的 `test_license.go` 测试程序：
```bash
# 确保已初始化系统（运行过 init 命令）
go run test_license.go
```
测试程序会自动：
1. 获取设备ID
2. 生成测试许可证
3. 保存到文件
4. 从文件读取并验证
5. 验证所有步骤是否成功

### Q: 如果用户替换了密钥文件会怎样？

A: 如果用户运行 `init` 生成新的密钥对，会覆盖原来的密钥文件。这会导致：
- ✅ 用户可以用新密钥生成自己的许可证
- ❌ 用户无法验证之前用旧密钥生成的许可证（签名不匹配）
- ❌ 用户无法破解其他用户的许可证（需要对应的私钥）
- ⚠️ 这是离线验证的固有特性，不是安全漏洞

**建议：**
- 对于高安全要求，使用在线验证模式
- 密钥文件应该妥善保管，不要随意替换
- 如果需要替换密钥，需要重新生成所有许可证

## 许可证

本项目采用 MIT 许可证。

## 贡献

欢迎提交 Issue 和 Pull Request。

## 作者

LicenseManager Team

