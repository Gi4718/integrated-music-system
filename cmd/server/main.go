package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"endfield-music/internal/api"
	"endfield-music/internal/config"
	"endfield-music/internal/db"

	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	port        int
	httpsServer *http.Server
	httpsPort   int
	mu          sync.Mutex
)

var rootCmd = &cobra.Command{
	Use:   "endfield-music-server",
	Short: "集成音乐系统 Web 服务",
	Long:  `集成音乐系统 Web 服务，提供 API 接口和静态文件服务`,
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "/data/config/config.yaml", "配置文件路径")
	rootCmd.Flags().IntVarP(&port, "port", "p", 33550, "HTTP 端口")
}

// loadTLSCert 加载 TLS 证书（支持中间证书链合并）
func loadTLSCert(certPath, keyPath, chainPath string) (tls.Certificate, error) {
	certFile := certPath
	if chainPath != "" {
		certData, err1 := os.ReadFile(certPath)
		chainData, err2 := os.ReadFile(chainPath)
		if err1 == nil && err2 == nil {
			fullChainPath := "/tmp/fullchain.pem"
			combined := append(certData, chainData...)
			if err := os.WriteFile(fullChainPath, combined, 0644); err == nil {
				certFile = fullChainPath
			}
		}
	}
	return tls.LoadX509KeyPair(certFile, keyPath)
}

// ReloadSSL 热加载 SSL 证书
func ReloadSSL() bool {
	mu.Lock()
	defer mu.Unlock()

	sslMode, _ := db.GetSetting("ssl_mode")
	sslCertPath, _ := db.GetSetting("ssl_cert_path")
	sslKeyPath, _ := db.GetSetting("ssl_key_path")
	sslChainPath, _ := db.GetSetting("ssl_chain_path")

	if sslMode == "none" || sslCertPath == "" || sslKeyPath == "" {
		log.Printf("SSL 未配置 (mode=%s, cert=%s, key=%s)", sslMode, sslCertPath, sslKeyPath)
		return false
	}

	cert, err := loadTLSCert(sslCertPath, sslKeyPath, sslChainPath)
	if err != nil {
		errMsg := fmt.Sprintf("SSL 证书加载失败: %v", err)
		log.Printf("警告: %s", errMsg)
		db.SetSetting("ssl_error", errMsg)
		db.SetSetting("ssl_redirect", "false")
		return false
	}

	httpsAddr := fmt.Sprintf(":%d", httpsPort)

	if httpsServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		httpsServer.Shutdown(ctx)
		log.Printf("HTTPS 服务器已关闭，准备重新启动")
	}

	httpsServer = &http.Server{
		Addr:    httpsAddr,
		Handler: api.GetRouter(),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		},
	}

	go func() {
		log.Printf("HTTPS 服务器启动于 %s", httpsAddr)
		db.SetSetting("ssl_error", "")
		api.SetHTTPSReady()
		if err := httpsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTPS 服务器错误: %v", err)
			api.ClearHTTPSReady()
		}
	}()

	return true
}

func startServer() {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		log.Printf("警告: 加载配置失败，使用默认配置: %v", err)
		cfg = config.DefaultConfig()
	}

	if err := db.InitDB("/data/db/netmusic.db"); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer db.CloseDB()

	router := api.SetupRouter(cfg)

	// 设置 SSL 热加载回调
	api.ReloadSSLFunc = ReloadSSL

	httpPort := port
	httpsPort = 33551

	// 从数据库读取端口配置
	if v, _ := db.GetSetting("http_port"); v != "" {
		fmt.Sscanf(v, "%d", &httpPort)
	}
	if v, _ := db.GetSetting("https_port"); v != "" {
		fmt.Sscanf(v, "%d", &httpsPort)
	}

	// 始终启动 HTTP 服务器
	httpAddr := fmt.Sprintf(":%d", httpPort)
	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: router,
	}

	go func() {
		log.Printf("HTTP 服务器启动于 %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP 服务器错误: %v", err)
		}
	}()

	// 检查 SSL 设置，决定是否启动 HTTPS 服务器
	sslMode, _ := db.GetSetting("ssl_mode")
	sslCertPath, _ := db.GetSetting("ssl_cert_path")
	sslKeyPath, _ := db.GetSetting("ssl_key_path")
	sslChainPath, _ := db.GetSetting("ssl_chain_path")

	if sslMode != "none" && sslCertPath != "" && sslKeyPath != "" {
		httpsAddr := fmt.Sprintf(":%d", httpsPort)

		// 尝试加载证书并启动 HTTPS
		startHTTPS := func() bool {
			cert, err := loadTLSCert(sslCertPath, sslKeyPath, sslChainPath)
			if err != nil {
				errMsg := fmt.Sprintf("SSL 证书加载失败: %v", err)
				log.Printf("警告: %s", errMsg)
				// 保存错误信息到数据库，供设置页显示
				db.SetSetting("ssl_error", errMsg)
				// 确保 HTTP 跳转关闭（保留 HTTP 模式）
				db.SetSetting("ssl_redirect", "false")
				return false
			}

			httpsServer := &http.Server{
				Addr:    httpsAddr,
				Handler: router,
				TLSConfig: &tls.Config{
					Certificates: []tls.Certificate{cert},
					MinVersion:   tls.VersionTLS12,
				},
			}

			go func() {
				log.Printf("HTTPS 服务器启动于 %s", httpsAddr)
				// 清除之前的 SSL 错误信息
				db.SetSetting("ssl_error", "")
				// 标记 HTTPS 已就绪
				api.SetHTTPSReady()
				if err := httpsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
					log.Printf("HTTPS 服务器错误: %v", err)
					api.ClearHTTPSReady()
				}
			}()
			return true
		}

		if !startHTTPS() {
			// 证书加载失败，启动后台定时重试
			go func() {
				ticker := time.NewTicker(30 * time.Second)
				defer ticker.Stop()
				for range ticker.C {
					if startHTTPS() {
						log.Printf("SSL 证书加载成功，HTTPS 服务已启动")
						return
					}
				}
			}()
		}
	} else {
		log.Printf("SSL 未启用 (mode=%s, cert=%s, key=%s)", sslMode, sslCertPath, sslKeyPath)
	}

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	httpServer.Shutdown(ctx)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
