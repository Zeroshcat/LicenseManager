package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Zeroshcat/LicenseManager/internal/database"
	"github.com/Zeroshcat/LicenseManager/internal/server"
)

const (
	defaultDBPath = "license.db"
	defaultPort   = 8080
)

func main() {
	// 解析命令行参数
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	host := fs.String("host", "0.0.0.0", "监听地址")
	port := fs.Int("port", defaultPort, "监听端口")
	dbPath := fs.String("db", defaultDBPath, "数据库文件路径")
	fs.Parse(os.Args[1:])

	// 创建数据库连接
	fmt.Printf("正在连接数据库: %s\n", *dbPath)
	db, err := database.NewDB(*dbPath)
	if err != nil {
		log.Fatalf("❌ 连接数据库失败: %v\n", err)
	}
	defer db.Close()
	fmt.Println("✅ 数据库连接成功")

	// 创建HTTP服务器
	httpServer := server.NewServer(db)

	// 设置路由
	addr := fmt.Sprintf("%s:%d", *host, *port)
	fmt.Printf("启动授权服务器: http://%s\n", addr)
	fmt.Println("API端点:")
	fmt.Printf("  - 健康检查: GET  http://%s/api/health\n", addr)
	fmt.Printf("  - 在线验证: POST http://%s/api/v1/license/verify/online\n", addr)
	fmt.Printf("  - 双重验证: POST http://%s/api/v1/license/verify/dual\n", addr)
	fmt.Printf("  - 设备注册: POST http://%s/api/v1/device/register\n", addr)
	fmt.Printf("  - 获取设备: GET  http://%s/api/v1/device/{device_id}\n", addr)
	fmt.Println()
	fmt.Println("按 Ctrl+C 停止服务器")

	// 设置优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 启动HTTP服务器（在goroutine中）
	go func() {
		if err := http.ListenAndServe(addr, httpServer); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ 服务器启动失败: %v\n", err)
		}
	}()

	// 等待中断信号
	<-sigChan
	fmt.Println("\n正在关闭服务器...")
	fmt.Println("✅ 服务器已停止")
}

