package api

import (
	"endfield-music/internal/config"
	"endfield-music/internal/db"
	"endfield-music/internal/service"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// httpsReady 标记 HTTPS 服务器是否已就绪（证书加载成功并启动）
var httpsReady atomic.Bool

// SetHTTPSReady 标记 HTTPS 已就绪
func SetHTTPSReady() {
	httpsReady.Store(true)
}

// ClearHTTPSReady 清除 HTTPS 就绪标记
func ClearHTTPSReady() {
	httpsReady.Store(false)
}

// IsHTTPSReady 查询 HTTPS 是否已就绪
func IsHTTPSReady() bool {
	return httpsReady.Load()
}

// ReloadSSLFunc 是 SSL 热加载回调函数，由 cmd/server/main.go 设置
var ReloadSSLFunc func() bool

// GetRouter 返回已配置的 router（用于热加载时获取新的 handler）
var currentRouter *gin.Engine

// GetRouter 返回当前 router
func GetRouter() *gin.Engine {
	return currentRouter
}

func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()
	currentRouter = router

	// 初始化任务服务（需在下载引擎之前创建，因为引擎需要引用它）
	taskService := service.NewTaskService()
	taskHandler := NewTaskHandler(taskService)

	// 初始化下载引擎（传入 taskService）
	InitDownloadEngine(cfg, taskService)

	// 健康检查
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// HTTP 到 HTTPS 重定向中间件
	router.Use(func(c *gin.Context) {
		if c.Request.TLS == nil {
			sslRedirect, _ := db.GetSetting("ssl_redirect")
			sslMode, _ := db.GetSetting("ssl_mode")

			// 只有 HTTPS 已就绪时才执行重定向
			if sslRedirect == "true" && sslMode != "none" && IsHTTPSReady() {
				certPath, _ := db.GetSetting("ssl_cert_path")
				keyPath, _ := db.GetSetting("ssl_key_path")

				if certPath != "" && keyPath != "" {
					cert, err := service.ParseCertificate(certPath)
					if err == nil && cert != nil {
						now := time.Now()
						if !now.Before(cert.NotBefore) && !now.After(cert.NotAfter) {
							httpsPort, _ := db.GetSetting("https_port")
							if httpsPort == "" {
								httpsPort = "33551"
							}

							host := c.Request.Host
							for i := len(host) - 1; i >= 0; i-- {
								if host[i] == ':' {
									host = host[:i]
									break
								}
							}

							targetURL := fmt.Sprintf("https://%s:%s%s", host, httpsPort, c.Request.RequestURI)
							c.Redirect(http.StatusMovedPermanently, targetURL)
							c.Abort()
							return
						}
					}
				}
			}
		}
		c.Next()
	})

	// API 路由组
	api := router.Group("/api")
	{
		// 系统认证路由（无需认证中间件）
		api.POST("/system/register", Register)
		api.POST("/system/login", SystemLogin)
		api.GET("/system/check", CheckSystemUser)

		// 网易云认证路由（无需 JWT 中间件，使用数据库 cookie 认证）
		auth := api.Group("/auth")
		{
			auth.GET("/qr-key", getQRKey)
			auth.GET("/qr-code", getQRCode)
			auth.GET("/qr-check", checkQRStatus)
			auth.GET("/status", getLoginStatus)
			auth.POST("/logout", logout)
			auth.POST("/sms/send", sendSMSCode)
			auth.POST("/phone", loginByPhone)
			auth.POST("/phone/password", loginByPhonePassword)
			auth.POST("/email", loginByEmail)
			auth.POST("/save-cookie", saveCookie)
			auth.POST("/second-verify", secondVerify)
			auth.POST("/qq", loginByQQ)
			auth.GET("/qq/url", getQQAuthURL)
		}

		// 设置路由（无需 JWT 中间件，使用数据库 cookie 认证）
		settings := api.Group("/settings")
		{
			settings.GET("", getSettings)
			settings.POST("", updateSettings)
			settings.POST("/ssl/upload", uploadSSLCert)
			settings.POST("/ssl/upload-file", uploadSSLCertFile)
			settings.POST("/ssl/validate", validateSSLCert)
			settings.POST("/ssl/acme", applyACME)
			settings.GET("/ssl/acme-plugins", getACMEPlugins)
		}

		// 需要 JWT 认证的路由组
		authorized := api.Group("")
		authorized.Use(AuthMiddleware())
		{
			// 推荐接口
			recommend := authorized.Group("/recommend")
			{
				recommend.GET("/songs", getRecommendSongs)
				recommend.GET("/playlists", getRecommendPlaylists)
			}

			search := authorized.Group("/search")
			{
				search.GET("/songs", searchSongs)
			}

			download := authorized.Group("/download")
			{
				download.POST("/song", downloadSong)
				download.POST("/playlist", downloadPlaylist)
				download.POST("/verify-metadata", verifyMetadata)
				download.GET("/history", getDownloadHistory)
				download.GET("/progress", getDownloadProgress)
			}

			playlist := authorized.Group("/playlist")
			{
				playlist.GET("/user", getUserPlaylists)
				playlist.GET("/detail", getPlaylistDetail)
				playlist.POST("/subscribe", subscribePlaylist)
			}

			player := authorized.Group("/player")
			{
				player.GET("/check/:id", checkSongURL)
				player.GET("/stream/:id", streamAudio)
			}

			tasks := authorized.Group("/tasks")
			{
				tasks.GET("", taskHandler.GetTasks)
				tasks.GET("/:id/progress", taskHandler.GetTaskProgress)
				tasks.POST("/:id/cancel", taskHandler.CancelTask)
			}
		}
	}

// 静态文件服务（Vue 前端）
	router.Static("/assets", "./web/dist/assets")
	router.StaticFile("/icon-black.ico", "./web/dist/icon-black.ico")
	router.StaticFile("/icon-light.ico", "./web/dist/icon-light.ico")
	router.StaticFile("/", "./web/dist/index.html")
	router.NoRoute(func(c *gin.Context) {
		c.File("./web/dist/index.html")
	})

	return router
}
