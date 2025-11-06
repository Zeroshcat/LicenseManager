// Package server 提供网络授权服务器功能
package server

import (
	"encoding/json"
	"net/http"
	"time"
	
	"github.com/Zeroshcat/LicenseManager/internal/database"
	"github.com/Zeroshcat/LicenseManager/pkg/license"
)

// Server 授权服务器
type Server struct {
	db      *database.DB
	handler http.Handler
}

// NewServer 创建授权服务器
// 参数：
//   - db: 数据库连接
// 返回值：
//   - *Server: 服务器实例
func NewServer(db *database.DB) *Server {
	s := &Server{db: db}
	s.setupRoutes()
	return s
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	mux := http.NewServeMux()
	
	// 健康检查
	mux.HandleFunc("/api/health", s.handleHealth)
	
	// 许可证验证端点
	mux.HandleFunc("/api/v1/license/verify/online", s.handleVerifyOnline)
	mux.HandleFunc("/api/v1/license/verify/dual", s.handleVerifyDual)
	
	// 设备管理端点
	mux.HandleFunc("/api/v1/device/register", s.handleRegisterDevice)
	mux.HandleFunc("/api/v1/device/", s.handleGetDevice)
	
	s.handler = mux
}

// ServeHTTP 实现http.Handler接口
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

// handleHealth 处理健康检查请求
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	response := map[string]interface{}{
		"status":    "ok",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

// handleVerifyOnline 处理网络许可证验证请求
func (s *Server) handleVerifyOnline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// 解析请求
	var req struct {
		DeviceID string `json:"device_id"`
		AppID    string `json:"app_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}
	
	// 查询许可证
	licenseRecord, err := s.db.GetLicenseByDeviceID(req.DeviceID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "LICENSE_NOT_FOUND", "License not found")
		return
	}
	
	// 检查是否过期
	now := time.Now()
	expired := now.After(licenseRecord.ExpiryDate)
	
	result := license.VerifyResult{
		Valid:       !expired,
		Expired:     expired,
		ExpiryDate:  licenseRecord.ExpiryDate,
		DeviceID:    req.DeviceID,
		LicenseType: licenseRecord.LicenseType,
		Message:     "Online verification",
	}
	
	if expired {
		result.Message = "License expired"
		s.writeJSON(w, http.StatusOK, result)
		return
	}
	
	s.writeJSON(w, http.StatusOK, result)
}

// handleVerifyDual 处理双重验证请求
func (s *Server) handleVerifyDual(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// 解析请求
	var req struct {
		LicenseKey string `json:"license_key"`
		DeviceID   string `json:"device_id"`
		AppID      string `json:"app_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}
	
	// 这里应该调用离线验证器验证licenseKey
	// 简化处理，只做网络验证部分
	_, err := s.db.GetDeviceByID(req.DeviceID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "DEVICE_NOT_FOUND", "Device not found")
		return
	}
	
	licenseRecord, err := s.db.GetLicenseByDeviceID(req.DeviceID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "LICENSE_NOT_FOUND", "License not found")
		return
	}
	
	now := time.Now()
	expired := now.After(licenseRecord.ExpiryDate)
	
	result := license.VerifyResult{
		Valid:        !expired,
		Expired:      expired,
		ExpiryDate:   licenseRecord.ExpiryDate,
		DeviceID:     req.DeviceID,
		LicenseType:  "dual",
		OfflineValid: true, // 应该从离线验证器获取
		OnlineValid:  !expired,
		Message:      "Dual verification",
	}
	
	s.writeJSON(w, http.StatusOK, result)
}

// handleRegisterDevice 处理设备注册请求
func (s *Server) handleRegisterDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		DeviceID   string `json:"device_id"`
		DeviceName string `json:"device_name"`
		AppID      string `json:"app_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}
	
	// 创建设备记录
	deviceRecord := &database.DeviceRecord{
		DeviceID:   req.DeviceID,
		DeviceName: req.DeviceName,
		AppID:      req.AppID,
		Status:     "active",
	}
	
	id, err := s.db.SaveDevice(deviceRecord)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "SERVER_ERROR", "Failed to register device")
		return
	}
	
	response := map[string]interface{}{
		"id":        id,
		"device_id": req.DeviceID,
		"status":    "registered",
	}
	
	s.writeJSON(w, http.StatusCreated, response)
}

// handleGetDevice 处理获取设备信息请求
func (s *Server) handleGetDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// 从URL提取device_id
	deviceID := r.URL.Path[len("/api/v1/device/"):]
	if deviceID == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Device ID required")
		return
	}
	
	deviceRecord, err := s.db.GetDeviceByID(deviceID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "DEVICE_NOT_FOUND", "Device not found")
		return
	}
	
	// 获取许可证信息
	licenseRecord, _ := s.db.GetLicenseByDeviceID(deviceID)
	
	response := map[string]interface{}{
		"device_id":     deviceRecord.DeviceID,
		"device_name":   deviceRecord.DeviceName,
		"registered_at": deviceRecord.RegisteredAt.Format(time.RFC3339),
		"last_seen":     deviceRecord.LastSeen.Format(time.RFC3339),
		"status":        deviceRecord.Status,
	}
	
	if licenseRecord != nil {
		response["license_status"] = "active"
		response["expiry_date"] = licenseRecord.ExpiryDate.Format(time.RFC3339)
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

// writeJSON 写入JSON响应
func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError 写入错误响应
func (s *Server) writeError(w http.ResponseWriter, status int, code, message string) {
	response := map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	}
	s.writeJSON(w, status, response)
}

