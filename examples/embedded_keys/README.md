# 嵌入密钥示例

这个示例演示如何将公钥和 AES 密钥嵌入到 Go 程序中，并使用代码混淆工具保护。

## 方法1：使用 embed 指令（推荐）

### 步骤1：准备密钥文件

**重要**：编译前必须将密钥文件复制到示例目录，否则编译会失败。

将密钥文件复制到示例目录：

```bash
# Linux/macOS
cp public_key.pem examples/embedded_keys/
cp aes_key.bin examples/embedded_keys/

# Windows
copy public_key.pem examples\embedded_keys\
copy aes_key.bin examples\embedded_keys\
```

### 步骤2：编译程序

```bash
cd examples/embedded_keys
go build -o app main.go
```

### 步骤3：使用 garble 混淆编译（可选但推荐）

安装 garble：

```bash
go install mvdan.cc/garble@latest
```

使用 garble 混淆编译：

```bash
garble build -o app main.go
```

**garble 的作用：**
- 混淆变量名和函数名
- 混淆字符串常量（包括嵌入的密钥）
- 增加逆向工程难度
- 减小二进制文件大小

### 步骤4：运行程序

```bash
./app
```

## 方法2：直接嵌入为字节数组（更安全但更复杂）

如果需要更高的安全性，可以将密钥转换为字节数组直接嵌入：

```go
package main

import (
    "github.com/Zeroshcat/LicenseManager/pkg/license"
)

// 嵌入的公钥（Base64编码后嵌入）
var embeddedPublicKey = []byte(`-----BEGIN PUBLIC KEY-----
...你的公钥内容...
-----END PUBLIC KEY-----`)

// 嵌入的AES密钥（32字节）
var embeddedAESKey = []byte{
    0x01, 0x02, 0x03, // ... 32个字节
}

func main() {
    verifier, err := license.NewOfflineVerifier(embeddedPublicKey, embeddedAESKey)
    // ... 使用验证器
}
```

## 方法3：使用 XOR 加密（额外保护层）

可以先用 XOR 加密密钥，然后在运行时解密：

```go
package main

import (
    "github.com/Zeroshcat/LicenseManager/pkg/license"
)

// XOR 加密的密钥（使用简单的XOR加密）
var encryptedPublicKey = []byte{/* 加密后的公钥 */}
var encryptedAESKey = []byte{/* 加密后的AES密钥 */}
var xorKey = []byte("your-secret-xor-key")

func decryptXOR(data []byte, key []byte) []byte {
    result := make([]byte, len(data))
    for i := range data {
        result[i] = data[i] ^ key[i%len(key)]
    }
    return result
}

func main() {
    // 解密密钥
    publicKey := decryptXOR(encryptedPublicKey, xorKey)
    aesKey := decryptXOR(encryptedAESKey, xorKey)
    
    verifier, err := license.NewOfflineVerifier(publicKey, aesKey)
    // ... 使用验证器
}
```

## 安全建议

1. **使用 garble 混淆**：强烈推荐使用 garble 工具混淆代码
2. **不要提交密钥文件**：确保 `.gitignore` 包含密钥文件
3. **定期轮换密钥**：定期更换密钥对
4. **使用在线验证**：对于高安全要求，使用在线验证模式

## 注意事项

- 即使使用混淆，密钥仍然可能被逆向工程提取
- 混淆只是增加难度，不是绝对安全
- 对于极高安全要求，建议使用在线验证模式

