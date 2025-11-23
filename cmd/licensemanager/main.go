package main

import (
	"crypto/rand"
	"crypto/rsa"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Zeroshcat/LicenseManager/internal/admin"
	"github.com/Zeroshcat/LicenseManager/internal/auth"
	"github.com/Zeroshcat/LicenseManager/internal/crypto"
	"github.com/Zeroshcat/LicenseManager/internal/database"
	licensegen "github.com/Zeroshcat/LicenseManager/internal/license"
	"github.com/Zeroshcat/LicenseManager/pkg/device"
	"github.com/Zeroshcat/LicenseManager/pkg/license"
	"github.com/Zeroshcat/LicenseManager/pkg/output"
)

const (
	defaultDBPath = "license.db"
	version       = "1.0.0"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "init":
		handleInit(args)
	case "generate":
		handleGenerate(args)
	case "verify":
		handleVerify(args)
	case "uuid", "checkuuid":
		handleUUID(args)
	case "device":
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "错误: device 命令需要子命令 (list/show/bind)\n")
			os.Exit(1)
		}
		handleDevice(args[0], args[1:])
	case "admin":
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "错误: admin 命令需要子命令 (serve/token)\n")
			os.Exit(1)
		}
		handleAdmin(args[0], args[1:])
	case "version":
		fmt.Printf("LicenseManager %s\n", version)
	default:
		fmt.Fprintf(os.Stderr, "错误: 未知命令 '%s'\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("LicenseManager - 统一许可证管理工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  licensemanager <command> [options]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  init             初始化数据库和生成密钥")
	fmt.Println("  generate         生成许可证")
	fmt.Println("  verify           验证许可证")
	fmt.Println("  uuid/checkuuid   获取设备UUID（用于生成许可证）")
	fmt.Println("  device           设备管理 (list/show/bind)")
	fmt.Println("  admin            管理功能 (serve/token)")
	fmt.Println("  version          显示版本信息")
	fmt.Println()
	fmt.Println("使用 'licensemanager <command> --help' 查看具体命令的帮助信息")
}

// handleInit 处理初始化命令
func handleInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	dbPath := fs.String("db", defaultDBPath, "数据库文件路径")
	fs.Parse(args)

	fmt.Println("正在初始化 LicenseManager...")

	// 1. 创建数据库
	fmt.Printf("创建数据库: %s\n", *dbPath)
	db, err := database.NewDB(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 创建数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	fmt.Println("✅ 数据库创建成功")

	// 2. 生成RSA密钥对
	fmt.Println("生成RSA密钥对...")
	privateKey, publicKey, err := crypto.GenerateRSAKeyPair()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 生成RSA密钥对失败: %v\n", err)
		os.Exit(1)
	}

	// 保存私钥
	privateKeyPEM := crypto.EncodePrivateKey(privateKey)
	if err := os.WriteFile("private_key.pem", privateKeyPEM, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 保存私钥失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ 私钥已保存: private_key.pem")

	// 保存公钥
	publicKeyPEM := crypto.EncodePublicKey(publicKey)
	if err := os.WriteFile("public_key.pem", publicKeyPEM, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 保存公钥失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ 公钥已保存: public_key.pem")

	// 3. 生成AES密钥
	fmt.Println("生成AES密钥...")
	aesKey := make([]byte, 32)
	if _, err := os.Stat("/dev/urandom"); err == nil {
		// Linux/macOS: 使用 /dev/urandom
		f, err := os.Open("/dev/urandom")
		if err == nil {
			f.Read(aesKey)
			f.Close()
		} else {
			rand.Read(aesKey)
		}
	} else {
		// Windows或其他系统: 使用crypto/rand
		rand.Read(aesKey)
	}

	if err := os.WriteFile("aes_key.bin", aesKey, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 保存AES密钥失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ AES密钥已保存: aes_key.bin")

	fmt.Println("\n初始化完成！")
	fmt.Println("⚠️  重要: 请妥善保管密钥文件，丢失后无法恢复！")
}

// handleGenerate 处理生成许可证命令
func handleGenerate(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	licenseType := fs.String("type", "offline", "许可证类型 (offline/online/dual)")
	deviceID := fs.String("device-id", "", "设备ID (留空则自动获取)")
	expiryDateStr := fs.String("expiry", "", "到期日期 (格式: YYYY-MM-DD)")
	outputFile := fs.String("output", "", "输出文件路径 (留空则输出到标准输出)")
	fs.Parse(args)

	// 验证必需参数
	if *expiryDateStr == "" {
		fmt.Fprintf(os.Stderr, "错误: 必须指定到期日期 (--expiry)\n")
		os.Exit(1)
	}

	// 解析到期日期
	expiryDate, err := time.Parse("2006-01-02", *expiryDateStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无效的日期格式: %v\n", err)
		os.Exit(1)
	}

	// 获取设备ID
	if *deviceID == "" {
		id, err := device.GetDeviceID()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 获取设备ID失败: %v\n", err)
			os.Exit(1)
		}
		*deviceID = id
		fmt.Printf("自动获取设备ID: %s\n", *deviceID)
	}

	// 验证许可证类型
	var licType license.LicenseType
	switch *licenseType {
	case "offline":
		licType = license.LicenseTypeOffline
	case "online":
		licType = license.LicenseTypeOnline
	case "dual":
		licType = license.LicenseTypeDual
	default:
		fmt.Fprintf(os.Stderr, "错误: 无效的许可证类型: %s\n", *licenseType)
		os.Exit(1)
	}

	// 加载密钥
	privateKey, aesKey, err := loadKeys()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 加载密钥失败: %v\n", err)
		os.Exit(1)
	}

	// 创建生成器
	generator := licensegen.NewGenerator(privateKey, aesKey)

	// 生成许可证
	licenseKey, err := generator.Generate(*deviceID, licType, expiryDate, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 生成许可证失败: %v\n", err)
		os.Exit(1)
	}

	// 保存到数据库
	db, err := database.NewDB(defaultDBPath)
	if err == nil {
		defer db.Close()
		licenseRecord := &database.LicenseRecord{
			DeviceID:    *deviceID,
			LicenseKey:  licenseKey,
			LicenseType: *licenseType,
			ExpiryDate:  expiryDate,
		}
		db.SaveLicense(licenseRecord)
	}

	// 输出许可证
	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, []byte(licenseKey), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存许可证文件失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ 许可证已保存到: %s\n", *outputFile)
	} else {
		fmt.Println(licenseKey)
	}
}

// handleVerify 处理验证许可证命令
func handleVerify(args []string) {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	licenseFile := fs.String("license-file", "", "许可证文件路径")
	deviceID := fs.String("device-id", "", "设备ID (留空则自动获取)")
	online := fs.Bool("online", false, "使用在线验证")
	dual := fs.Bool("dual", false, "使用双重验证")
	apiURL := fs.String("api-url", "", "API地址 (在线验证和双重验证需要)")
	fs.Parse(args)

	// 获取设备ID
	if *deviceID == "" {
		id, err := device.GetDeviceID()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 获取设备ID失败: %v\n", err)
			os.Exit(1)
		}
		*deviceID = id
	}

	var result *license.VerifyResult
	var err error

	if *dual {
		// 双重验证
		if *licenseFile == "" || *apiURL == "" {
			fmt.Fprintf(os.Stderr, "错误: 双重验证需要 --license-file 和 --api-url\n")
			os.Exit(1)
		}
		licenseKey, err := license.LoadLicenseFromFile(*licenseFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 读取许可证文件失败: %v\n", err)
			os.Exit(1)
		}
		publicKeyPEM, aesKey, err := loadPublicKeys()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 加载密钥失败: %v\n", err)
			os.Exit(1)
		}
		verifier, err := license.NewDualVerifier(&license.DualConfig{
			APIURL:  *apiURL,
			AppID:   "default",
			Timeout: 10,
		}, publicKeyPEM, aesKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 创建验证器失败: %v\n", err)
			os.Exit(1)
		}
		result, err = verifier.Verify(licenseKey, *deviceID)

	} else if *online {
		// 在线验证
		if *apiURL == "" {
			fmt.Fprintf(os.Stderr, "错误: 在线验证需要 --api-url\n")
			os.Exit(1)
		}
		verifier := license.NewOnlineVerifier(&license.OnlineConfig{
			APIURL:  *apiURL,
			AppID:   "default",
			Timeout: 10,
		})
		result, err = verifier.Verify(*deviceID)

	} else {
		// 离线验证
		if *licenseFile == "" {
			fmt.Fprintf(os.Stderr, "错误: 离线验证需要 --license-file\n")
			os.Exit(1)
		}
		licenseKey, err := license.LoadLicenseFromFile(*licenseFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 读取许可证文件失败: %v\n", err)
			os.Exit(1)
		}
		publicKeyPEM, aesKey, err := loadPublicKeys()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 加载密钥失败: %v\n", err)
			os.Exit(1)
		}
		verifier, err := license.NewOfflineVerifier(publicKeyPEM, aesKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 创建验证器失败: %v\n", err)
			os.Exit(1)
		}
		result, err = verifier.Verify(licenseKey, *deviceID)
	}

	if err != nil {
		if result != nil {
			fmt.Printf("❌ 验证失败: %v\n", err)
			fmt.Printf("   设备ID: %s\n", result.DeviceID)
			fmt.Printf("   到期时间: %s\n", result.ExpiryDate.Format("2006-01-02 15:04:05"))
			fmt.Printf("   消息: %s\n", result.Message)
		} else {
			fmt.Fprintf(os.Stderr, "❌ 验证失败: %v\n", err)
		}
		os.Exit(1)
	}

	if result.Valid && !result.Expired {
		fmt.Println("✅ 许可证验证成功！")
		fmt.Printf("   设备ID: %s\n", result.DeviceID)
		fmt.Printf("   许可证类型: %s\n", result.LicenseType)
		fmt.Printf("   到期时间: %s\n", result.ExpiryDate.Format("2006-01-02 15:04:05"))
		if result.OfflineValid || result.OnlineValid {
			fmt.Printf("   离线验证: %v\n", result.OfflineValid)
			fmt.Printf("   在线验证: %v\n", result.OnlineValid)
		}
	} else {
		fmt.Println("❌ 许可证无效或已过期")
		os.Exit(1)
	}
}

// handleUUID 处理获取设备UUID命令（参考 xinjiayu/LicenseManager）
func handleUUID(args []string) {
	deviceID, err := device.GetDeviceID()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 获取设备UUID失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(deviceID)
}

// handleDevice 处理设备管理命令
func handleDevice(subcmd string, args []string) {
	switch subcmd {
	case "list":
		handleDeviceList(args)
	case "show":
		handleDeviceShow(args)
	case "bind":
		handleDeviceBind(args)
	default:
		fmt.Fprintf(os.Stderr, "错误: 未知的设备命令 '%s'\n", subcmd)
		fmt.Println("可用命令: list/show/bind")
		os.Exit(1)
	}
}

func handleDeviceList(args []string) {
	fs := flag.NewFlagSet("device list", flag.ExitOnError)
	format := fs.String("format", "text", "输出格式 (text/json)")
	fs.Parse(args)

	db, err := database.NewDB(defaultDBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	devices, err := db.ListDevices(100, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 查询设备失败: %v\n", err)
		os.Exit(1)
	}

	formatter := output.GetFormatter(output.Format(*format))
	if *format == "json" {
		formatter.Print(devices)
	} else {
		if len(devices) == 0 {
			fmt.Println("没有找到设备")
			return
		}
		fmt.Println("设备列表:")
		for _, d := range devices {
			fmt.Printf("  ID: %d, DeviceID: %s, Name: %s, Status: %s\n",
				d.ID, d.DeviceID, d.DeviceName, d.Status)
		}
	}
}

func handleDeviceShow(args []string) {
	if len(args) == 0 {
		// 显示当前设备ID
		deviceID, err := device.GetDeviceID()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 获取设备ID失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("当前设备ID: %s\n", deviceID)
		return
	}

	deviceID := args[0]
	db, err := database.NewDB(defaultDBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	deviceRecord, err := db.GetDeviceByID(deviceID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 查询设备失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("设备信息:\n")
	fmt.Printf("  ID: %d\n", deviceRecord.ID)
	fmt.Printf("  DeviceID: %s\n", deviceRecord.DeviceID)
	fmt.Printf("  名称: %s\n", deviceRecord.DeviceName)
	fmt.Printf("  状态: %s\n", deviceRecord.Status)
	fmt.Printf("  注册时间: %s\n", deviceRecord.RegisteredAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  最后访问: %s\n", deviceRecord.LastSeen.Format("2006-01-02 15:04:05"))
}

func handleDeviceBind(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "错误: 需要指定设备ID\n")
		os.Exit(1)
	}

	deviceID := args[0]
	db, err := database.NewDB(defaultDBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	deviceRecord := &database.DeviceRecord{
		DeviceID:   deviceID,
		DeviceName: "Unknown",
		Status:     "active",
	}

	id, err := db.SaveDevice(deviceRecord)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 绑定设备失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 设备绑定成功 (ID: %d)\n", id)
}

// handleAdmin 处理管理命令
func handleAdmin(subcmd string, args []string) {
	switch subcmd {
	case "serve":
		handleAdminServe(args)
	case "token":
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "错误: token 命令需要子命令 (create)\n")
			os.Exit(1)
		}
		handleAdminToken(args[0], args[1:])
	default:
		fmt.Fprintf(os.Stderr, "错误: 未知的管理命令 '%s'\n", subcmd)
		fmt.Println("可用命令: serve/token")
		os.Exit(1)
	}
}

func handleAdminServe(args []string) {
	fs := flag.NewFlagSet("admin serve", flag.ExitOnError)
	password := fs.String("passwd", "", "管理密码 (必需)")
	host := fs.String("host", "0.0.0.0", "监听地址 (默认: 0.0.0.0)")
	port := fs.Int("port", 8080, "监听端口")
	dbPath := fs.String("db", defaultDBPath, "数据库文件路径")
	fs.Parse(args)

	if *password == "" {
		fmt.Fprintf(os.Stderr, "错误: 必须指定管理密码 (--passwd)\n")
		os.Exit(1)
	}

	// 创建数据库连接
	fmt.Printf("正在连接数据库: %s\n", *dbPath)
	db, err := database.NewDB(*dbPath)
	if err != nil {
		log.Fatalf("❌ 连接数据库失败: %v\n", err)
	}
	defer db.Close()
	fmt.Println("✅ 数据库连接成功")

	// 创建Web管理界面
	webAdmin, err := admin.NewWebAdmin(db, *password)
	if err != nil {
		log.Fatalf("❌ 创建管理界面失败: %v\n", err)
	}

	addr := fmt.Sprintf("%s:%d", *host, *port)
	fmt.Printf("启动Web管理服务器: http://%s\n", addr)
	fmt.Println("功能:")
	fmt.Println("  - 统计概览")
	fmt.Println("  - 设备管理")
	fmt.Println("  - 许可证管理（在线生成和下载）")
	fmt.Println("  - Token管理")
	fmt.Println()
	fmt.Println("按 Ctrl+C 停止服务器")

	// 设置优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 启动HTTP服务器（在goroutine中）
	go func() {
		if err := http.ListenAndServe(addr, webAdmin); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ 服务器启动失败: %v\n", err)
		}
	}()

	// 等待中断信号
	<-sigChan
	fmt.Println("\n正在关闭服务器...")
	fmt.Println("✅ 服务器已停止")
}

func handleAdminToken(subcmd string, args []string) {
	if subcmd != "create" {
		fmt.Fprintf(os.Stderr, "错误: 未知的token命令 '%s'\n", subcmd)
		os.Exit(1)
	}

	fs := flag.NewFlagSet("admin token create", flag.ExitOnError)
	tokenType := fs.String("type", "client", "Token类型 (client/admin)")
	appID := fs.String("app-id", "", "应用ID (client类型需要)")
	expiresStr := fs.String("expires", "", "过期日期 (格式: YYYY-MM-DD, 留空表示永不过期)")
	fs.Parse(args)

	if *tokenType == "client" && *appID == "" {
		fmt.Fprintf(os.Stderr, "错误: client类型需要指定 --app-id\n")
		os.Exit(1)
	}

	db, err := database.NewDB(defaultDBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// 生成Token
	token, err := auth.GenerateToken(32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 生成Token失败: %v\n", err)
		os.Exit(1)
	}

	// 解析过期时间
	var expiresAt *time.Time
	if *expiresStr != "" {
		expiry, err := time.Parse("2006-01-02", *expiresStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 无效的日期格式: %v\n", err)
			os.Exit(1)
		}
		expiresAt = &expiry
	}

	// 保存Token
	tokenRecord := &database.TokenRecord{
		Token:     token,
		TokenType: *tokenType,
		AppID:     *appID,
		ExpiresAt: expiresAt,
	}

	id, err := db.SaveToken(tokenRecord)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 保存Token失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Token创建成功 (ID: %d)\n", id)
	fmt.Printf("Token: %s\n", token)
	if expiresAt != nil {
		fmt.Printf("过期时间: %s\n", expiresAt.Format("2006-01-02"))
	}
}

// loadKeys 加载私钥和AES密钥
func loadKeys() (*rsa.PrivateKey, []byte, error) {
	privateKeyPEM, err := os.ReadFile("private_key.pem")
	if err != nil {
		return nil, nil, fmt.Errorf("读取私钥失败: %w", err)
	}

	privateKey, err := crypto.DecodePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("解码私钥失败: %w", err)
	}

	aesKey, err := os.ReadFile("aes_key.bin")
	if err != nil {
		return nil, nil, fmt.Errorf("读取AES密钥失败: %w", err)
	}

	if len(aesKey) != 32 {
		return nil, nil, fmt.Errorf("AES密钥长度不正确: 期望32字节，实际%d字节", len(aesKey))
	}

	return privateKey, aesKey, nil
}

// loadPublicKeys 加载公钥和AES密钥
func loadPublicKeys() ([]byte, []byte, error) {
	publicKeyPEM, err := os.ReadFile("public_key.pem")
	if err != nil {
		return nil, nil, fmt.Errorf("读取公钥失败: %w", err)
	}

	aesKey, err := os.ReadFile("aes_key.bin")
	if err != nil {
		return nil, nil, fmt.Errorf("读取AES密钥失败: %w", err)
	}

	if len(aesKey) != 32 {
		return nil, nil, fmt.Errorf("AES密钥长度不正确: 期望32字节，实际%d字节", len(aesKey))
	}

	return publicKeyPEM, aesKey, nil
}
