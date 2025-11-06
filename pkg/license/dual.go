// Package license 提供许可证生成和验证功能
package license

// DualConfig 双重验证配置
type DualConfig struct {
	APIURL  string // API地址（必须）
	AppID   string // 应用ID（必须）
	Timeout int    // 超时时间（秒）
}

// DualVerifier 双重验证器
// 需要同时满足离线验证和网络验证
type DualVerifier struct {
	offlineVerifier *OfflineVerifier // 离线验证器
	onlineVerifier  *OnlineVerifier  // 网络验证器
}

// NewDualVerifier 创建双重验证器
// 参数：
//   - config: 双重验证配置
//   - publicKeyPEM: RSA公钥（PEM格式，用于离线验证）
//   - aesKey: AES密钥（32字节，用于解密）
// 返回值：
//   - *DualVerifier: 双重验证器实例
func NewDualVerifier(config *DualConfig, publicKeyPEM []byte, aesKey []byte) *DualVerifier {
	// 创建离线验证器
	offlineVerifier := NewOfflineVerifier(publicKeyPEM, aesKey)
	
	// 创建网络验证器
	onlineConfig := &OnlineConfig{
		APIURL:  config.APIURL,
		AppID:   config.AppID,
		Timeout: config.Timeout,
	}
	onlineVerifier := NewOnlineVerifier(onlineConfig)
	
	return &DualVerifier{
		offlineVerifier: offlineVerifier,
		onlineVerifier:  onlineVerifier,
	}
}

// Verify 验证双重许可证
// 需要同时通过离线验证和网络验证
// 参数：
//   - licenseKey: 许可证密钥（用于离线验证）
//   - deviceID: 设备ID
// 返回值：
//   - *VerifyResult: 验证结果
//   - error: 验证过程中的错误
func (v *DualVerifier) Verify(licenseKey string, deviceID string) (*VerifyResult, error) {
	// 执行离线验证
	offlineResult, err := v.offlineVerifier.Verify(licenseKey, deviceID)
	if err != nil {
		return &VerifyResult{
			Valid:        false,
			OfflineValid: false,
			OnlineValid:  false,
			Message:      "Offline verification failed",
		}, err
	}
	
	// 执行网络验证
	onlineResult, err := v.onlineVerifier.Verify(deviceID)
	if err != nil {
		return &VerifyResult{
			Valid:        false,
			OfflineValid: offlineResult.Valid,
			OnlineValid:  false,
			Message:      "Online verification failed",
		}, err
	}
	
	// 两者都必须通过
	valid := offlineResult.Valid && onlineResult.Valid && !offlineResult.Expired && !onlineResult.Expired
	
	result := &VerifyResult{
		Valid:        valid,
		Expired:      offlineResult.Expired || onlineResult.Expired,
		ExpiryDate:   offlineResult.ExpiryDate,
		DeviceID:     deviceID,
		LicenseType:  "dual",
		OfflineValid: offlineResult.Valid && !offlineResult.Expired,
		OnlineValid:  onlineResult.Valid && !onlineResult.Expired,
		Message:      "Dual verification",
	}
	
	if !valid {
		result.Message = "Dual verification failed"
		return result, ErrInvalidLicense
	}
	
	return result, nil
}

