// Package admin 提供后台管理功能
package admin

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Zeroshcat/LicenseManager/internal/crypto"
	"github.com/Zeroshcat/LicenseManager/internal/database"
	licensegen "github.com/Zeroshcat/LicenseManager/internal/license"
	"github.com/Zeroshcat/LicenseManager/pkg/license"
)

// WebAdmin Web管理界面
type WebAdmin struct {
	db       *database.DB
	template *template.Template
	password string // 管理密码
}

// NewWebAdmin 创建Web管理界面
// 参数：
//   - db: 数据库连接
//   - password: 管理密码
//
// 返回值：
//   - *WebAdmin: Web管理界面实例
func NewWebAdmin(db *database.DB, password string) (*WebAdmin, error) {
	admin := &WebAdmin{
		db:       db,
		password: password,
	}

	// 从文件加载HTML模板
	// 尝试多个可能的路径
	templatePaths := []string{
		"web/index.html",
		"./web/index.html",
		filepath.Join(filepath.Dir(os.Args[0]), "web", "index.html"),
	}

	var tmpl *template.Template
	var err error
	for _, path := range templatePaths {
		if _, err := os.Stat(path); err == nil {
			tmpl, err = template.ParseFiles(path)
			if err == nil {
				break
			}
		}
	}

	// 如果所有路径都失败，使用内嵌的简单模板
	if tmpl == nil {
		tmpl, err = template.New("admin").Parse(simpleTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template: %w", err)
		}
	}

	admin.template = tmpl

	return admin, nil
}

// ServeHTTP 实现http.Handler接口
func (w *WebAdmin) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// 密码验证（除了登录接口）
	if r.URL.Path != "/api/login" {
		if !w.checkAuth(r) {
			if r.Method == http.MethodGet && (r.URL.Path == "/" || r.URL.Path == "/index.html") {
				// 首页需要登录，重定向到登录页面
				w.handleLogin(rw, r)
				return
			}
			// API请求返回401
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// 路由处理
	switch r.URL.Path {
	case "/", "/index.html":
		w.handleIndex(rw, r)
	case "/api/login":
		w.handleLoginAPI(rw, r)
	case "/api/stats":
		w.handleStatsAPI(rw, r)
	case "/api/devices":
		w.handleDevicesAPI(rw, r)
	case "/api/licenses":
		w.handleLicensesAPI(rw, r)
	case "/api/licenses/generate":
		w.handleGenerateLicense(rw, r)
	case "/api/tokens":
		w.handleTokensAPI(rw, r)
	default:
		path := r.URL.Path
		// 处理下载许可证文件: GET /api/licenses/{id}/download
		if r.Method == http.MethodGet && strings.HasPrefix(path, "/api/licenses/") && strings.HasSuffix(path, "/download") {
			w.handleDownloadLicense(rw, r)
		} else if r.Method == http.MethodDelete && strings.HasPrefix(path, "/api/licenses/") {
			// 删除许可证: DELETE /api/licenses/{id}
			// 确保不是下载路径
			if !strings.HasSuffix(path, "/download") {
				w.handleDeleteLicense(rw, r)
			} else {
				rw.Header().Set("Content-Type", "application/json")
				rw.WriteHeader(http.StatusMethodNotAllowed)
				json.NewEncoder(rw).Encode(map[string]interface{}{
					"success": false,
					"message": "Method not allowed",
				})
			}
		} else if r.Method == http.MethodPost && strings.HasSuffix(path, "/revoke") {
			// 撤销Token: POST /api/tokens/{token}/revoke
			w.handleRevokeToken(rw, r)
		} else {
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusNotFound)
			json.NewEncoder(rw).Encode(map[string]interface{}{
				"success": false,
				"message": "Not found",
			})
		}
	}
}

// checkAuth 检查认证
func (w *WebAdmin) checkAuth(r *http.Request) bool {
	// 从Cookie获取密码
	cookie, err := r.Cookie("admin_auth")
	if err != nil {
		return false
	}
	// 简单的密码比较（实际应该使用哈希）
	return cookie.Value == w.password
}

// handleLogin 处理登录页面
func (w *WebAdmin) handleLogin(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(rw, loginPageHTML)
}

// handleLoginAPI 处理登录API
func (w *WebAdmin) handleLoginAPI(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(rw, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Password != w.password {
		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid password",
		})
		return
	}

	// 设置Cookie
	http.SetCookie(rw, &http.Cookie{
		Name:     "admin_auth",
		Value:    w.password,
		Path:     "/",
		MaxAge:   86400, // 24小时
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(map[string]interface{}{
		"success": true,
		"message": "Login successful",
	})
}

// handleIndex 处理首页
func (w *WebAdmin) handleIndex(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取统计数据
	stats, err := w.db.GetStats()
	if err != nil {
		stats = map[string]int64{
			"total_devices":    0,
			"active_devices":   0,
			"total_licenses":   0,
			"expired_licenses": 0,
		}
	}

	// 转换为interface{}类型供模板使用
	statsData := make(map[string]interface{})
	for k, v := range stats {
		statsData[k] = v
	}

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := w.template.Execute(rw, statsData); err != nil {
		http.Error(rw, "Failed to render template", http.StatusInternalServerError)
	}
}

// handleStatsAPI 处理统计数据API
func (w *WebAdmin) handleStatsAPI(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := w.db.GetStats()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(stats)
}

// handleDevicesAPI 处理设备API
func (w *WebAdmin) handleDevicesAPI(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit

	// 查询设备
	devices, err := w.db.ListDevices(limit, offset)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回JSON
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(map[string]interface{}{
		"devices": devices,
		"page":    page,
		"limit":   limit,
	})
}

// handleLicensesAPI 处理许可证API
func (w *WebAdmin) handleLicensesAPI(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}

	offset := (page - 1) * limit

	// 查询许可证
	licenses, err := w.db.ListLicenses(limit, offset)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回JSON
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(map[string]interface{}{
		"licenses": licenses,
		"page":     page,
		"limit":    limit,
	})
}

// handleDeleteLicense 处理删除许可证
func (w *WebAdmin) handleDeleteLicense(rw http.ResponseWriter, r *http.Request) {
	// 从URL提取ID: /api/licenses/{id}
	path := r.URL.Path
	prefix := "/api/licenses/"
	if !strings.HasPrefix(path, prefix) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid path",
		})
		return
	}

	idStr := strings.TrimPrefix(path, prefix)
	// 移除可能的尾部斜杠
	idStr = strings.TrimSuffix(idStr, "/")

	if idStr == "" {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(map[string]interface{}{
			"success": false,
			"message": "License ID is required",
		})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid license ID: " + err.Error(),
		})
		return
	}

	if err := w.db.DeleteLicense(id); err != nil {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(rw).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to delete license: " + err.Error(),
		})
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(map[string]interface{}{
		"success": true,
		"message": "License deleted successfully",
	})
}

// handleTokensAPI 处理Token API
func (w *WebAdmin) handleTokensAPI(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}

	offset := (page - 1) * limit

	// 查询Token
	tokens, err := w.db.ListTokens(limit, offset)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回JSON
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(map[string]interface{}{
		"tokens": tokens,
		"page":   page,
		"limit":  limit,
	})
}

// handleRevokeToken 处理撤销Token
func (w *WebAdmin) handleRevokeToken(rw http.ResponseWriter, r *http.Request) {
	// 从URL提取token
	// /api/tokens/{token}/revoke
	path := r.URL.Path
	token := path[len("/api/tokens/"):]
	token = token[:len(token)-7] // 移除 /revoke

	if err := w.db.RevokeToken(token); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(map[string]interface{}{
		"success": true,
		"message": "Token revoked",
	})
}

// handleGenerateLicense 处理生成许可证
func (w *WebAdmin) handleGenerateLicense(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DeviceID    string `json:"device_id"`
		LicenseType string `json:"license_type"`
		ExpiryDate  string `json:"expiry_date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(rw, "Invalid request", http.StatusBadRequest)
		return
	}

	// 验证必需参数
	if req.DeviceID == "" {
		http.Error(rw, "device_id is required", http.StatusBadRequest)
		return
	}
	if req.ExpiryDate == "" {
		http.Error(rw, "expiry_date is required", http.StatusBadRequest)
		return
	}

	// 解析到期时间
	expiryDate, err := time.Parse("2006-01-02", req.ExpiryDate)
	if err != nil {
		http.Error(rw, "Invalid expiry date format (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	// 加载密钥（需要从文件加载）
	privateKey, aesKey, err := w.loadKeys()
	if err != nil {
		http.Error(rw, "Failed to load keys: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 验证许可证类型
	var licType license.LicenseType
	switch req.LicenseType {
	case "offline":
		licType = license.LicenseTypeOffline
	case "online":
		licType = license.LicenseTypeOnline
	case "dual":
		licType = license.LicenseTypeDual
	default:
		http.Error(rw, "Invalid license type (offline|online|dual)", http.StatusBadRequest)
		return
	}

	// 创建生成器
	generator := licensegen.NewGenerator(privateKey, aesKey)

	// 生成许可证
	licenseKey, err := generator.Generate(req.DeviceID, licType, expiryDate, nil)
	if err != nil {
		http.Error(rw, "Failed to generate license: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 保存到数据库
	licenseRecord := &database.LicenseRecord{
		DeviceID:    req.DeviceID,
		LicenseKey:  licenseKey,
		LicenseType: req.LicenseType,
		ExpiryDate:  expiryDate,
	}

	_, err = w.db.SaveLicense(licenseRecord)
	if err != nil {
		http.Error(rw, "Failed to save license: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回JSON，包含许可证ID用于下载
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(map[string]interface{}{
		"success":     true,
		"license_key": licenseKey,
		"license_id":  licenseRecord.ID,
		"message":     "License generated successfully",
	})
}

// handleDownloadLicense 处理下载许可证文件
func (w *WebAdmin) handleDownloadLicense(rw http.ResponseWriter, r *http.Request) {
	// 从URL提取ID: /api/licenses/{id}/download
	path := r.URL.Path
	prefix := "/api/licenses/"
	suffix := "/download"

	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		http.Error(rw, "Invalid download path", http.StatusBadRequest)
		return
	}

	// 提取ID部分
	idStr := strings.TrimPrefix(path, prefix)
	idStr = strings.TrimSuffix(idStr, suffix)

	if idStr == "" {
		http.Error(rw, "License ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(rw, "Invalid license ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 获取许可证记录
	license, err := w.db.GetLicenseByID(id)
	if err != nil {
		http.Error(rw, "License not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// 设置下载响应头，文件名为 license.key
	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.Header().Set("Content-Disposition", "attachment; filename=license.key")
	rw.Header().Set("Content-Length", fmt.Sprintf("%d", len(license.LicenseKey)))

	// 写入许可证密钥内容（license_key 字段）
	rw.Write([]byte(license.LicenseKey))
}

// loadKeys 加载密钥文件
func (w *WebAdmin) loadKeys() (*rsa.PrivateKey, []byte, error) {
	// 尝试多个可能的路径
	keyPaths := []string{
		"private_key.pem",
		"./private_key.pem",
		filepath.Join(filepath.Dir(os.Args[0]), "private_key.pem"),
	}

	var privateKeyPEM []byte
	var readErr error
	for _, path := range keyPaths {
		if _, statErr := os.Stat(path); statErr == nil {
			privateKeyPEM, readErr = os.ReadFile(path)
			if readErr == nil {
				break
			}
		}
	}
	if readErr != nil || privateKeyPEM == nil {
		return nil, nil, fmt.Errorf("failed to read private key")
	}

	privateKey, err := crypto.DecodePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	// 加载AES密钥
	aesPaths := []string{
		"aes_key.bin",
		"./aes_key.bin",
		filepath.Join(filepath.Dir(os.Args[0]), "aes_key.bin"),
	}

	var aesKey []byte
	var aesReadErr error
	for _, path := range aesPaths {
		if _, statErr := os.Stat(path); statErr == nil {
			aesKey, aesReadErr = os.ReadFile(path)
			if aesReadErr == nil {
				break
			}
		}
	}
	if aesReadErr != nil || aesKey == nil {
		return nil, nil, fmt.Errorf("failed to read AES key")
	}

	return privateKey, aesKey, nil
}

// loginPageHTML 登录页面HTML
const loginPageHTML = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LicenseManager - 登录</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
        }
        .login-container {
            background: white;
            padding: 2.5rem;
            border-radius: 12px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            width: 100%;
            max-width: 400px;
        }
        .login-container h1 {
            color: #2c3e50;
            margin-bottom: 0.5rem;
            font-size: 1.8rem;
        }
        .login-container p {
            color: #7f8c8d;
            margin-bottom: 2rem;
        }
        .form-group {
            margin-bottom: 1.5rem;
        }
        .form-group label {
            display: block;
            color: #2c3e50;
            margin-bottom: 0.5rem;
            font-weight: 500;
        }
        .form-group input {
            width: 100%;
            padding: 0.75rem;
            border: 2px solid #e0e0e0;
            border-radius: 6px;
            font-size: 1rem;
            transition: border-color 0.3s;
        }
        .form-group input:focus {
            outline: none;
            border-color: #667eea;
        }
        .btn-login {
            width: 100%;
            padding: 0.75rem;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 6px;
            font-size: 1rem;
            font-weight: 600;
            cursor: pointer;
            transition: background 0.3s;
        }
        .btn-login:hover {
            background: #5568d3;
        }
        .error {
            color: #e74c3c;
            margin-top: 0.5rem;
            font-size: 0.9rem;
        }
    </style>
</head>
<body>
    <div class="login-container">
        <h1>LicenseManager</h1>
        <p>请输入管理密码以继续</p>
        <form id="loginForm">
            <div class="form-group">
                <label for="password">密码</label>
                <input type="password" id="password" name="password" required autofocus>
                <div id="error" class="error" style="display: none;"></div>
            </div>
            <button type="submit" class="btn-login">登录</button>
        </form>
    </div>
    <script>
        document.getElementById('loginForm').addEventListener('submit', function(e) {
            e.preventDefault();
            const password = document.getElementById('password').value;
            const errorDiv = document.getElementById('error');
            
            fetch('/api/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ password: password })
            })
            .then(res => res.json())
            .then(data => {
                if (data.success) {
                    window.location.href = '/';
                } else {
                    errorDiv.textContent = data.message || '登录失败';
                    errorDiv.style.display = 'block';
                }
            })
            .catch(err => {
                errorDiv.textContent = '登录失败，请重试';
                errorDiv.style.display = 'block';
            });
        });
    </script>
</body>
</html>
`

// simpleTemplate 简单的内嵌模板（作为后备）
const simpleTemplate = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>LicenseManager - 管理后台</title>
    <style>
        body { font-family: sans-serif; padding: 2rem; }
        .error { color: red; }
    </style>
</head>
<body>
    <h1>LicenseManager 管理后台</h1>
    <p class="error">无法加载模板文件，请确保 web/index.html 文件存在。</p>
    <p>统计数据：</p>
    <ul>
        <li>总设备数: {{.total_devices}}</li>
        <li>活跃设备: {{.active_devices}}</li>
        <li>总许可证: {{.total_licenses}}</li>
        <li>已过期: {{.expired_licenses}}</li>
    </ul>
</body>
</html>
`
