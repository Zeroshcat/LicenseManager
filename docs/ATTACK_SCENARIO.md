# 攻击场景分析：替换密钥对能否破解许可证？

## 攻击场景

**用户B 的操作：**
1. 运行 `licensemanager init` 生成新的密钥对
2. 这会生成：
   - 新的 `private_key.pem`（用户B的私钥）
   - 新的 `public_key.pem`（用户B的公钥）
   - 新的 `aes_key.bin`（用户B的AES密钥）
3. 用户B 用这些新密钥替换了原来的文件
4. 用户B 尝试验证用户A 的许可证

## 许可证生成和验证流程

### 生成流程（用户A）：
```
1. 创建许可证对象（JSON）
   ↓
2. AES加密（使用用户A的AES密钥）
   ↓
3. RSA签名（使用用户A的私钥）
   ↓
4. Base64编码
   ↓
5. 生成 license.key
```

### 验证流程（用户B尝试）：
```
1. Base64解码
   ↓
2. 提取签名和密文
   ↓
3. 验证RSA签名（使用用户B的公钥）❌ 失败！
   ↓
4. 如果签名验证失败，直接返回错误，不会继续解密
```

## 关键点分析

### 为什么无法破解？

**原因1：RSA签名验证失败**

许可证的签名是用**用户A的私钥**生成的，只能用**用户A的公钥**验证。

如果用户B用自己生成的公钥去验证用户A的许可证签名：
- 签名验证会失败（因为公钥和私钥不匹配）
- 验证流程会在签名验证阶段就失败
- 根本不会执行到AES解密步骤

**代码验证：**

```go
// pkg/license/offline.go:136
valid, err := crypto.VerifySignature(encryptedData, signature, v.publicKey)
if err != nil || !valid {
    return nil, ErrInvalidLicense  // 直接返回错误，不会继续
}
```

**结论：** 即使用户B替换了公钥和AES密钥，签名验证也会失败，无法破解。

### 如果用户B用自己的密钥生成新许可证呢？

**用户B 可以做什么：**
1. ✅ 用自己的私钥生成新的许可证
2. ✅ 用自己的公钥验证自己生成的许可证
3. ❌ 无法验证用户A的许可证（签名不匹配）
4. ❌ 无法破解用户A的许可证

**用户B 无法做什么：**
- ❌ 无法用自己生成的密钥验证用户A的许可证
- ❌ 无法伪造用户A的许可证
- ❌ 无法修改用户A的许可证（签名会失效）

## 实际测试

让我们模拟这个攻击场景：

### 场景1：用户A 生成许可证

```bash
# 用户A
./licensemanager init
# 生成：private_key_A.pem, public_key_A.pem, aes_key_A.bin

./licensemanager generate --type offline \
  --device-id device123 \
  --expiry 2026-12-31
# 生成许可证，使用 private_key_A.pem 和 aes_key_A.bin
```

### 场景2：用户B 替换密钥

```bash
# 用户B
./licensemanager init
# 生成：private_key_B.pem, public_key_B.pem, aes_key_B.bin
# 这会覆盖原来的 public_key_A.pem 和 aes_key_A.bin

# 用户B 尝试验证用户A的许可证
./licensemanager verify --license-file license.key --device-id device123
```

### 结果：验证失败 ❌

**错误信息：** `invalid license` 或 `failed to decode license`

**原因：**
- 用户B的公钥无法验证用户A的许可证签名
- 签名验证失败，返回 `ErrInvalidLicense`
- 不会继续执行AES解密

## 安全结论

### ✅ 无法破解的原因：

1. **RSA签名验证是第一步**
   - 验证流程首先检查签名
   - 签名验证失败，直接返回错误
   - 不会执行后续的AES解密

2. **公钥和私钥必须匹配**
   - 用户A的许可证用用户A的私钥签名
   - 只能用用户A的公钥验证
   - 用户B的公钥无法验证用户A的签名

3. **即使替换了AES密钥也没用**
   - 因为签名验证已经失败
   - 根本不会执行到AES解密步骤

### ⚠️ 需要注意的情况：

**如果用户B同时拥有：**
- 用户A的公钥（`public_key_A.pem`）
- 用户A的AES密钥（`aes_key_A.bin`）
- 用户A的许可证文件

**那么用户B可以：**
- ✅ 验证用户A的许可证（签名验证通过）
- ✅ 解密许可证内容（查看设备ID、过期时间等）
- ❌ 无法修改许可证（需要用户A的私钥重新签名）
- ❌ 无法伪造新许可证（需要用户A的私钥）

**但是：**
- 如果用户B替换了公钥和AES密钥，就无法验证用户A的许可证了
- 因为签名验证会失败

## 总结

**回答你的问题：**

> 用户运行 `licensemanager init` 生成新的密钥对，替换了公钥和aes_key.bin，不就可以破解吗？

**答案：不能破解** ❌

**原因：**
1. 用户B生成的新密钥对与用户A的密钥对不匹配
2. 用户A的许可证是用用户A的私钥签名的
3. 只能用用户A的公钥验证签名
4. 用户B的公钥无法验证用户A的许可证签名
5. 签名验证失败，直接返回错误，不会继续解密

**关键点：**
- RSA签名验证是验证流程的第一步
- 签名验证失败，整个验证流程就失败了
- 即使用户B替换了AES密钥，也无法破解，因为签名验证已经失败

**真正的安全威胁：**
- 如果用户B同时拥有用户A的公钥和AES密钥，可以验证用户A的许可证
- 但即使这样，也无法修改或伪造许可证（需要私钥）

