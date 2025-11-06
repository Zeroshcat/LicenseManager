// Package license 提供许可证生成和验证功能
package license

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OnlineConfig 网络验证配置
type OnlineConfig struct {
	APIURL  string // API地址（必须）
	AppID   string // 应用ID（必须）
	Timeout int    // 超时时间（秒）
	Retries int    // 重试次数
}

// OnlineVerifier 网络验证器
// 需要预设API地址，通过网络验证
type OnlineVerifier struct {
	config   *OnlineConfig
	client   *http.Client
}

// NewOnlineVerifier 创建网络验证器
// 参数：
//   - config: 网络验证配置
// 返回值：
//   - *OnlineVerifier: 网络验证器实例
func NewOnlineVerifier(config *OnlineConfig) *OnlineVerifier {
	// 设置默认值
	if config.Timeout == 0 {
		config.Timeout = 10
	}
	if config.Retries == 0 {
		config.Retries = 3
	}
	
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}
	
	return &OnlineVerifier{
		config: config,
		client: client,
	}
}

// Verify 验证网络许可证
// 参数：
//   - deviceID: 设备ID
// 返回值：
//   - *VerifyResult: 验证结果
//   - error: 验证过程中的错误
func (v *OnlineVerifier) Verify(deviceID string) (*VerifyResult, error) {
	// 构建请求
	reqBody := map[string]string{
		"device_id": deviceID,
		"app_id":    v.config.AppID,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	
	// 发送请求
	url := fmt.Sprintf("%s/license/verify/online", v.config.APIURL)
	resp, err := v.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, ErrNetworkError
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrNetworkError
	}
	
	// 解析响应
	var result VerifyResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, ErrNetworkError
	}
	
	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, ErrNetworkError
	}
	
	return &result, nil
}


