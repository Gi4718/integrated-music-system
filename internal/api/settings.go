package api

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"endfield-music/internal/db"
	"endfield-music/internal/service"

	"github.com/gin-gonic/gin"
)

func getSettings(c *gin.Context) {
	// 获取当前系统用户 ID
	systemUserID := 0
	if id, exists := c.Get("system_user_id"); exists {
		if uid, ok := id.(int); ok {
			systemUserID = uid
		}
	}

	settings := map[string]interface{}{
		"download_path":   getSettingByUser(systemUserID, "download_path", ""),
		"song_format":     getSettingByUser(systemUserID, "song_format", "{songName} - {artist}"),
		"quality":         getSettingByUser(systemUserID, "quality", "high"),
		"auto_sync":       getSettingByUser(systemUserID, "auto_sync", "false"),
		"sync_interval":   getSettingByUser(systemUserID, "sync_interval", "12"),
		"sync_unit":       getSettingByUser(systemUserID, "sync_unit", "hour"),
		"delete_removed":  getSettingByUser(systemUserID, "delete_removed", "false"),
		"playlist_format": getSettingByUser(systemUserID, "playlist_format", "{playlistName}/{songName} - {artist}"),
		"resume_downloads": getSettingByUser(systemUserID, "resume_downloads", "true"),
		"auto_data_complete":      getSettingByUser(systemUserID, "auto_data_complete", "false"),
		"data_complete_interval":  getSettingByUser(systemUserID, "data_complete_interval", "24"),
		"data_complete_unit":      getSettingByUser(systemUserID, "data_complete_unit", "hour"),
		"data_complete_cover":     getSettingByUser(systemUserID, "data_complete_cover", "true"),
		"data_complete_lyrics":    getSettingByUser(systemUserID, "data_complete_lyrics", "true"),
		"data_complete_artist":    getSettingByUser(systemUserID, "data_complete_artist", "true"),
		"ssl_mode":        getSetting("ssl_mode", "none"),
		"ssl_cert_path":   getSetting("ssl_cert_path", ""),
		"ssl_key_path":    getSetting("ssl_key_path", ""),
		"ssl_chain_path":  getSetting("ssl_chain_path", ""),
		"http_port":       getSetting("http_port", "33550"),
		"https_port":      getSetting("https_port", "33551"),
		"ssl_redirect":    getSetting("ssl_redirect", "false"),
		"acme_provider":   getSetting("acme_provider", ""),
		"acme_email":      getSetting("acme_email", ""),
		"acme_domain":     getSetting("acme_domain", ""),
		"acme_account_id": getSetting("acme_account_id", ""),
		"acme_secret_key": getSetting("acme_secret_key", ""),
		"acme_token":      getSetting("acme_token", ""),
		"acme_region_id":  getSetting("acme_region_id", ""),
		"last_sync_time":  getSettingByUser(systemUserID, "last_sync_time", ""),
		"next_sync_time":  getSettingByUser(systemUserID, "next_sync_time", ""),
		"disable_page_animation": getSettingByUser(systemUserID, "disable_page_animation", "false"),
		"multi_user_enabled":      getSetting("multi_user_enabled", "false"),
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

func updateSettings(c *gin.Context) {
	// 获取当前系统用户 ID
	systemUserID := 0
	if id, exists := c.Get("system_user_id"); exists {
		if uid, ok := id.(int); ok {
			systemUserID = uid
		}
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 用户级别的设置（每个用户独立）
	userSettings := []string{
		"download_path", "song_format", "quality", "auto_sync", "sync_interval", "sync_unit",
		"delete_removed", "playlist_format", "resume_downloads", "auto_data_complete",
		"data_complete_interval", "data_complete_unit", "data_complete_cover",
		"data_complete_lyrics", "data_complete_artist", "last_sync_time", "next_sync_time",
		"disable_page_animation",
	}

	// 全局设置（所有用户共享）
	globalSettings := []string{
		"ssl_mode", "ssl_cert_path", "ssl_key_path", "ssl_chain_path",
		"http_port", "https_port", "ssl_redirect", "acme_provider", "acme_email",
		"acme_domain", "acme_account_id", "acme_secret_key", "acme_token", "acme_region_id",
		"multi_user_enabled",
	}

	for key, value := range req {
		strValue := fmt.Sprintf("%v", value)
		
		// 判断是用户级别还是全局设置
		isUserSetting := false
		for _, userKey := range userSettings {
			if key == userKey {
				isUserSetting = true
				break
			}
		}
		
		if isUserSetting {
			db.SetSettingByUser(systemUserID, key, strValue)
		} else {
			// 检查是否是全局设置
			isGlobalSetting := false
			for _, globalKey := range globalSettings {
				if key == globalKey {
					isGlobalSetting = true
					break
				}
			}
			if isGlobalSetting {
				db.SetSetting(key, strValue)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "设置已保存"})
}

func uploadSSLCert(c *gin.Context) {
	certFile, err := c.FormFile("cert")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传证书文件"})
		return
	}

	keyFile, err := c.FormFile("key")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传密钥文件"})
		return
	}

	certPath := filepath.Join("/data/ssl", "cert.pem")
	keyPath := filepath.Join("/data/ssl", "key.pem")

	if err := c.SaveUploadedFile(certFile, certPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存证书失败"})
		return
	}

	if err := c.SaveUploadedFile(keyFile, keyPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存密钥失败"})
		return
	}

	db.SetSetting("ssl_enabled", "true")
	db.SetSetting("ssl_mode", "upload")
	db.SetSetting("ssl_cert_file", certPath)
	db.SetSetting("ssl_key_file", keyPath)

	c.JSON(http.StatusOK, gin.H{
		"message":   "SSL 证书已上传",
		"cert_path": certPath,
		"key_path":  keyPath,
	})
}

func uploadSSLCertFile(c *gin.Context) {
	fileType := c.PostForm("type")
	if fileType == "" {
		fileType = "cert"
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传文件"})
		return
	}

	var filename string
	switch fileType {
	case "cert":
		filename = "cert.pem"
	case "key":
		filename = "key.pem"
	case "chain":
		filename = "chain.pem"
	default:
		filename = "cert.pem"
	}

	containerPath := filepath.Join("/data/ssl", filename)

	if err := c.SaveUploadedFile(file, containerPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "文件已上传",
		"container_path": containerPath,
	})
}

func validateSSLCert(c *gin.Context) {
	var req struct {
		CertPath string `json:"cert_path" binding:"required"`
		KeyPath  string `json:"key_path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	cert, err := service.ParseCertificate(req.CertPath)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"valid":   false,
			"message": "证书解析失败: " + err.Error(),
		})
		return
	}

	now := time.Now()
	if now.Before(cert.NotBefore) {
		c.JSON(http.StatusOK, gin.H{
			"valid":   false,
			"message": "证书尚未生效",
		})
		return
	}

	if now.After(cert.NotAfter) {
		c.JSON(http.StatusOK, gin.H{
			"valid":   false,
			"message": "证书已过期",
		})
		return
	}

	keyData, err := os.ReadFile(req.KeyPath)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"valid":   false,
			"message": "无法读取私钥文件",
		})
		return
	}

	keyBlock, _ := pem.Decode(keyData)
	if keyBlock == nil {
		c.JSON(http.StatusOK, gin.H{
			"valid":   false,
			"message": "私钥解析失败",
		})
		return
	}

	keyMatch := false
	switch pub := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		if privKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes); err == nil {
			keyMatch = privKey.PublicKey.N.Cmp(pub.N) == 0
		} else if privKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes); err == nil {
			if rsaKey, ok := privKey.(*rsa.PrivateKey); ok {
				keyMatch = rsaKey.PublicKey.N.Cmp(pub.N) == 0
			}
		}
	case *ecdsa.PublicKey:
		if privKey, err := x509.ParseECPrivateKey(keyBlock.Bytes); err == nil {
			keyMatch = privKey.PublicKey.X.Cmp(pub.X) == 0 && privKey.PublicKey.Y.Cmp(pub.Y) == 0
		} else if privKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes); err == nil {
			if ecKey, ok := privKey.(*ecdsa.PrivateKey); ok {
				keyMatch = ecKey.PublicKey.X.Cmp(pub.X) == 0 && ecKey.PublicKey.Y.Cmp(pub.Y) == 0
			}
		}
	}

	if !keyMatch {
		c.JSON(http.StatusOK, gin.H{
			"valid":   false,
			"message": "私钥与证书不匹配",
		})
		return
	}

	domains := []string{}
	if cert.Subject.CommonName != "" {
		domains = append(domains, cert.Subject.CommonName)
	}
	domains = append(domains, cert.DNSNames...)

	c.JSON(http.StatusOK, gin.H{
		"valid":      true,
		"message":    "证书有效",
		"subject":    cert.Subject.CommonName,
		"issuer":     cert.Issuer.CommonName,
		"not_before": cert.NotBefore.Format("2006-01-02"),
		"not_after":  cert.NotAfter.Format("2006-01-02"),
		"domains":    domains,
	})
}

func applyACME(c *gin.Context) {
	var req struct {
		Provider string            `json:"provider" binding:"required"`
		Email    string            `json:"email" binding:"required"`
		Domain   string            `json:"domain" binding:"required"`
		Fields   map[string]string `json:"fields"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if req.Fields == nil {
		req.Fields = make(map[string]string)
	}

	config := service.ACMEConfig{
		Provider: req.Provider,
		Email:    req.Email,
		Domain:   req.Domain,
		Fields:   req.Fields,
	}

	configJSON, _ := json.Marshal(config)
	db.SetSetting("acme_config", string(configJSON))
	db.SetSetting("ssl_mode", "acme")

	certPath, keyPath, err := service.RunACME(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("证书申请失败: %v", err)})
		return
	}

	db.SetSetting("ssl_cert_path", certPath)
	db.SetSetting("ssl_key_path", keyPath)

	c.JSON(http.StatusOK, gin.H{
		"message":   "证书申请成功",
		"cert_path": certPath,
		"key_path":  keyPath,
	})
}

func getACMEPlugins(c *gin.Context) {
	plugins := service.GetPlugins()
	c.JSON(http.StatusOK, gin.H{"plugins": plugins})
}

func getSetting(key, defaultValue string) string {
	value, err := db.GetSetting(key)
	if err != nil || value == "" {
		return defaultValue
	}
	return value
}

func getSettingByUser(userID int, key, defaultValue string) string {
	value, err := db.GetSettingByUser(userID, key)
	if err != nil || value == "" {
		return defaultValue
	}
	return value
}
