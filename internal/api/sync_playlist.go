package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"endfield-music/internal/db"

	"github.com/gin-gonic/gin"
)

// GetSyncPlaylists 获取同步歌单列表
func GetSyncPlaylists(c *gin.Context) {
	userID := getCurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	playlists, err := db.GetSyncPlaylists(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取同步歌单失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    playlists,
	})
}

// UpdateSyncPlaylists 更新同步歌单列表
func UpdateSyncPlaylists(c *gin.Context) {
	userID := getCurrentUserID(c)
	fmt.Printf("[UpdateSyncPlaylists] userID=%d\n", userID)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		PlaylistIDs []int `json:"playlist_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[UpdateSyncPlaylists] bind error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误", "details": err.Error()})
		return
	}

	fmt.Printf("[UpdateSyncPlaylists] playlistIDs=%v\n", req.PlaylistIDs)

	if err := db.BatchUpdateSyncPlaylists(userID, req.PlaylistIDs); err != nil {
		fmt.Printf("[UpdateSyncPlaylists] batch update error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新同步歌单失败", "details": err.Error()})
		return
	}

	fmt.Printf("[UpdateSyncPlaylists] success\n")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "同步歌单已更新",
	})
}

// ToggleSyncPlaylist 切换歌单同步状态
func ToggleSyncPlaylist(c *gin.Context) {
	userID := getCurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		PlaylistID int  `json:"playlist_id"`
		Enabled    bool `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := db.ToggleSyncPlaylist(req.PlaylistID, userID, req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "切换同步状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "同步状态已更新",
	})
}

// TriggerManualSync 手动触发同步
func TriggerManualSync(c *gin.Context) {
	// 获取下载引擎
	eng := getDownloadEngine()
	if eng == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "下载引擎未初始化"})
		return
	}

	// 获取同步模式
	syncMode, _ := db.GetSetting("sync_mode")

	// 如果是间隔模式，重置计时器
	if syncMode != "schedule" {
		now := time.Now()
		db.SetSetting("last_sync_time", now.Format(time.RFC3339))

		// 计算下次同步时间
		intervalStr, _ := db.GetSetting("sync_interval")
		unitStr, _ := db.GetSetting("sync_unit")
		interval := 12
		if intervalStr != "" {
			fmt.Sscanf(intervalStr, "%d", &interval)
		}
		if interval <= 0 {
			interval = 12
		}

		if unitStr == "day" {
			interval *= 24
		}

		nextTime := now.Add(time.Duration(interval) * time.Hour)
		db.SetSetting("next_sync_time", nextTime.Format(time.RFC3339))
	}

	// 异步执行同步，避免阻塞HTTP请求
	go func() {
		eng.RunAutoSync(context.Background())
	}()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "同步任务已启动",
	})
}

// getCurrentUserID 获取当前登录用户ID
func getCurrentUserID(c *gin.Context) int {
	// 优先从 system_user_id 获取（OptionalAuthMiddleware 设置的）
	if userID, exists := c.Get("system_user_id"); exists {
		if id, ok := userID.(int); ok && id > 0 {
			return id
		}
	}
	// 兼容 user_id
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(int); ok && id > 0 {
			return id
		}
	}
	// Fallback: 查询第一个系统用户
	return getSystemUserID(c)
}
